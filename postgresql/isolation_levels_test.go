package postgresql

import (
	"context"
	"fmt"
	"licklib/utils"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CREATE USER lick WITH PASSWORD 'lick';
// CREATE DATABASE licklib WITH OWNER lick;

func connect(ctx context.Context, t *testing.T) *pgxpool.Pool {
	conn, err := pgxpool.New(
		ctx,
		fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			"localhost", 5432, "lick", "lick", "licklib",
		),
	)
	require.NoError(t, err)
	require.NoError(t, conn.Ping(ctx))
	return conn
}

// TestLostUpdate демонстрирует аномалию "потерянного обновления" данных при уровне изоляции READ COMMITTED,
// а также её устранение использованием REPEATABLE READ.
func TestLostUpdate(t *testing.T) {
	ctx := context.Background()
	conn := connect(ctx, t)
	defer conn.Close()

	testFlow := func(repeatableRead bool) {
		// Создаём тестовую таблицу: ID пользователя и его баланс.
		tableName := utils.RandomString(32)
		_, err := conn.Exec(
			ctx,
			fmt.Sprintf(
				`
			CREATE TABLE %v
			(
				id TEXT PRIMARY KEY,
				balance INTEGER
			);
			`,
				tableName,
			),
		)
		defer dropTable(ctx, t, conn, tableName)
		require.NoError(t, err)

		// Сохраняем исходные данные в таблице: у пользователя "Боб" баланс 300 руб.
		_, err = conn.Exec(ctx, fmt.Sprintf(`INSERT INTO %v VALUES ($1, $2);`, tableName), "bob", 300)
		require.NoError(t, err)

		// Начинаем транзакцию №1.
		// В зависимости от сценария проставляем уровень изоляции транзакции №1: READ COMMITTED или REPEATABLE READ.
		opts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
		if repeatableRead {
			opts.IsoLevel = pgx.RepeatableRead
		}
		tx1, err := conn.BeginTx(ctx, opts)
		require.NoError(t, err)
		// Считываем баланс Боба (300 руб.).
		balanceTx1 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&balanceTx1))
		assert.Equal(t, 300, balanceTx1)

		// Начинаем транзакцию №2.
		// Уровень изоляции транзакции №2 есть READ COMMITTED, т.е. по умолчанию.
		tx2, err := conn.Begin(ctx)
		require.NoError(t, err)
		// Считываем баланс Боба (300 руб.) и отнимаем от него 100 руб.
		// Итого в рамках транзакции №2 у Боба осталось 200 руб.
		balanceTx2 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&balanceTx2))
		_, err = tx2.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), balanceTx2-100, "bob")
		require.NoError(t, err)
		// Завершаем транзакцию №2.
		require.NoError(t, tx2.Commit(ctx))

		// В рамках транзакции №1 пополняем баланс Боба на 500 руб., используя ранее считанное значение баланса.
		_, err = tx1.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), balanceTx1+500, "bob")
		if repeatableRead {
			// При уровне изоляции REPEATABLE READ произойдет ошибка сериализации,
			// так как данные в строке были уже изменены в транзакции №2.
			assert.Error(t, err)
		} else {
			// При уровне изоляции READ COMMITTED ошибки не произойдёт,
			// и обновление баланса, совершённое в рамках транзакции №2, будет потеряно.
			assert.NoError(t, err)
		}
		// Завершаем транзакцию №1.
		err = tx1.Commit(ctx)
		if repeatableRead {
			// При ошибке на предыдущем шаге на уровне изоляции REPEATABLE READ правильнее было бы сразу сделать tx1.Rollback.
			// При tx1.Commit завершение транзакции произойдёт с ошибкой "unexpected rollback".
			require.Error(t, err)
		} else {
			// При уровне изоляции READ COMMITTED транзакция №1 будет завершена без ошибки.
			require.NoError(t, err)
		}

		// Проверяем итоговый баланс Боба.
		actualBalance := 0
		require.NoError(t, conn.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&actualBalance))
		if repeatableRead {
			// При уровне изоляции REPEATABLE READ будут сохранены изменения, внесённые только транзакцией №2.
			// Баланс Боба был уменьшен на 100 руб. и составляет 200 руб. Получилось избежать грубой ошибки потери обновления.
			assert.Equal(t, 300-100, actualBalance)
		} else {
			// При уровне изоляции READ COMMITTED обновление транзакции №2 будет потеряно (аномалия),
			// так как не была выдана ошибка при повторном обновлении строки в рамках транзакции №1.
			// Такие образом транзакция №1 произвела обновление, основанное на "старом" чтении, произошедшем до действий транзакции №2.
			// Баланс Боба был увеличен на 500 руб. и составляет 800 руб., что является грубой ошибкой.
			assert.Equal(t, 300+500, actualBalance)
		}
	}

	t.Run("anomaly on READ COMMITTED", func(t *testing.T) { testFlow(false) })
	t.Run("proper behavior on REPEATABLE READ", func(t *testing.T) { testFlow(true) })
}

// TestNonRepeatableRead демонстрирует аномалию "неповторяющегося чтения" данных при уровне изоляции READ COMMITTED,
// а также её устранение использованием REPEATABLE READ.
func TestNonRepeatableRead(t *testing.T) {
	ctx := context.Background()
	conn := connect(ctx, t)
	defer conn.Close()

	testFlow := func(repeatableRead bool) {
		// Создаём тестовую таблицу: ID пользователя и его баланс.
		tableName := utils.RandomString(32)
		_, err := conn.Exec(
			ctx,
			fmt.Sprintf(
				`
			CREATE TABLE %v
			(
				id TEXT PRIMARY KEY,
				balance INTEGER
			);
			`,
				tableName,
			),
		)
		defer dropTable(ctx, t, conn, tableName)
		require.NoError(t, err)

		// Сохраняем исходные данные в таблице: у пользователя "Боб" баланс 300 руб.
		_, err = conn.Exec(ctx, fmt.Sprintf(`INSERT INTO %v VALUES ($1, $2);`, tableName), "bob", 300)
		require.NoError(t, err)

		// Начинаем транзакцию №1.
		// В зависимости от сценария проставляем уровень изоляции транзакции №1: READ COMMITTED или REPEATABLE READ.
		opts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
		if repeatableRead {
			opts.IsoLevel = pgx.RepeatableRead
		}
		tx1, err := conn.BeginTx(ctx, opts)
		require.NoError(t, err)
		// Считываем баланс Боба (300 руб.).
		initBalanceTx1 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&initBalanceTx1))
		assert.Equal(t, 300, initBalanceTx1)
		// Рассчитаем доход по вкладу на основе считанного значения.
		income := initBalanceTx1 / 10

		// Начинаем транзакцию №2.
		// Уровень изоляции транзакции №2 есть READ COMMITTED, т.е. по умолчанию.
		tx2, err := conn.Begin(ctx)
		require.NoError(t, err)
		// Считываем баланс Боба (300 руб.) и пополняем на 100 руб.
		// Итого в рамках транзакции №2 у Боба осталось 200 руб.
		balanceTx2 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&balanceTx2))
		_, err = tx2.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), balanceTx2+100, "bob")
		require.NoError(t, err)
		// Завершаем транзакцию №2.
		require.NoError(t, tx2.Commit(ctx))

		// В рамках транзакции №1 вновь считываем баланс Боба.
		balanceTx1 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&balanceTx1))
		if repeatableRead {
			// При уровне изоляции REPEATABLE READ изменения, произведённые в рамках транзакции №2,
			// не повлияют на повторное чтение в транзакции №1. Баланс будет по-прежнему 300 руб.
			assert.Equal(t, 300, balanceTx1)
		} else {
			// При уровне изоляции READ COMMITTED повторное чтение после завершённой транзакции №2
			// даст уже новое значение баланса в 400 руб., что есть "неповторяющееся чтение" (аномалия).
			assert.Equal(t, 400, balanceTx1)
		}

		// В рамках транзакции №1 пополняем баланс Боба на сумму дохода по вкладу, рассчитанную при первом чтении.
		_, err = tx1.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), balanceTx1+income, "bob")
		if repeatableRead {
			// При уровне изоляции REPEATABLE READ произойдет ошибка сериализации,
			// так как данные в строке были уже изменены в транзакции №2.
			assert.Error(t, err)
		} else {
			// При уровне изоляции READ COMMITTED ошибки не произойдёт,
			// и доход от вклада будет в итоге спорным.
			assert.NoError(t, err)
		}
		// Завершаем транзакцию №1.
		err = tx1.Commit(ctx)
		if repeatableRead {
			// При ошибке на предыдущем шаге на уровне изоляции REPEATABLE READ правильнее было бы сразу сделать tx1.Rollback.
			// При tx1.Commit завершение транзакции произойдёт с ошибкой "unexpected rollback".
			require.Error(t, err)
		} else {
			// При уровне изоляции READ COMMITTED транзакция №1 будет завершена без ошибки.
			require.NoError(t, err)
		}

		// Проверяем итоговый баланс Боба.
		actualBalance := 0
		require.NoError(t, conn.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&actualBalance))
		if repeatableRead {
			// При уровне изоляции REPEATABLE READ будут сохранены изменения, внесённые только транзакцией №2.
			// Баланс Боба был увеличен на 100 руб. и составляет 400 руб. Получилось избежать грубой ошибки неповторяющегося чтпния.
			// Транзакция по начислению дохода будет запущена повторно, и доход будет рассчитан от верной суммы, без потерь.
			assert.Equal(t, 300+100, actualBalance)
		} else {
			// При уровне изоляции READ COMMITTED доход от вклада будет рассчитан при первом чтении (от 300 руб.).
			// Однако конкурирующая транзакция повысила баланс на 100 руб., вследствие чего сумма дохода является спорной:
			// не понятно, от 300 или 400 рублей её нужно рассчитывать. Будет сохранён один из вариантов, который может вызвать споры.
			// Имеем дело с аномалией неповторяющегося чтения.
			assert.Equal(t, 400+30, actualBalance)
		}
	}
	t.Run("anomaly on READ COMMITTED", func(t *testing.T) { testFlow(false) })
	t.Run("proper behavior on REPEATABLE READ", func(t *testing.T) { testFlow(true) })
}

// TestNonRepeatableRead демонстрирует аномалию "фантомного чтения" данных при уровне изоляции REPEATABLE READ,
// а также её устранение использованием SERIALIZABLE.
// NOTE: оказалось очень сложно "обмануть" PostgreSQL и привести её в неконсистентное состояние.
func TestPhantomRead(t *testing.T) {
	ctx := context.Background()
	conn := connect(ctx, t)
	defer conn.Close()

	testFlow := func(serializable bool) {
		// Создаём тестовую таблицу: ID пользователя и его баланс.
		tableName := utils.RandomString(32)
		_, err := conn.Exec(
			ctx,
			fmt.Sprintf(
				`
			CREATE TABLE %v
			(
				id TEXT PRIMARY KEY,
				balance INTEGER
			);
			`,
				tableName,
			),
		)
		defer dropTable(ctx, t, conn, tableName)
		require.NoError(t, err)

		// Сохраняем исходные данные в таблице.
		_, err = conn.Exec(
			ctx,
			fmt.Sprintf(`INSERT INTO %v VALUES ($1, $2), ($3, $4), ($5, $6);`, tableName),
			"bob",
			300,
			"pit",
			1000,
			"tor",
			1500,
		)
		require.NoError(t, err)

		// Начинаем транзакцию №1.
		// В зависимости от сценария проставляем уровень изоляции транзакции №1: REPEATABLE READ или SERIALIZABLE.
		opts := pgx.TxOptions{IsoLevel: pgx.RepeatableRead}
		if serializable {
			opts.IsoLevel = pgx.Serializable
		}
		tx1, err := conn.BeginTx(ctx, opts)
		require.NoError(t, err)

		// Начинаем транзакцию №2.
		// Уровень изоляции транзакции №2 есть READ COMMITTED, т.е. по умолчанию.
		tx2, err := conn.Begin(ctx)
		require.NoError(t, err)

		// В рамках транзакции №1 считываем количество пользователей, их 3.
		users := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %v;`, tableName)).Scan(&users))
		assert.Equal(t, 3, users)
		// За каждого мы должны получить по 100 руб в качестве реферальных.
		income := users * 100

		// В момент расчётов привели ещё 3000 человек. Записываем их в рамках транзакции №2.
		for i := range 3000 {
			_, err = tx2.Exec(ctx, fmt.Sprintf(`INSERT INTO %v VALUES ($1, $2);`, tableName), fmt.Sprint(i), i*1000)
			require.NoError(t, err)
		}
		// Завершаем транзакцию №2.
		require.NoError(t, tx2.Commit(ctx))

		// В рамках транзакции №1 считываем количество пользователей, баланс которых достиг 1000 руб.
		richUsers := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %v WHERE balance >= 1000;`, tableName)).Scan(&richUsers))
		// Каждому такому пользователю купим подарок за 10 руб.
		spends := richUsers * 10

		// Завершаем транзакцию №2.
		require.NoError(t, tx1.Commit(ctx))

		if serializable {
			// При уровне изоляции SERIALIZABLE мы купим подарок только тем, кто был в базе изначально,
			// т.е. двоим людям с балансом от 1000.
			assert.Equal(t, 2*10, spends)
		} else {
			// В рамках транзакции №1 (REPEATABLE READ) случилось "фантомное чтение" добавленных в транзакции №2 пользователей.
			// Это вызвало неверный расчёт стоимости подароков.
			// NOTE: не удалось вызвать аномалию "фантомного чтения" на PostgreSQL. Он очень умён.
			// assert.Equal(t, 3002*10, spends)
		}
		// Доход от рефералок рассчитан от 3-х пользователей. Стоимость подарков должна быть определена среди них же.
		assert.Equal(t, 3*100, income)
	}
	t.Run("anomaly on REPEATABLE READ", func(t *testing.T) { testFlow(false) })
	t.Run("proper behavior on SERIALIZABLE", func(t *testing.T) { testFlow(true) })
}

// TestDirtyRead демонстрирует аномалию грязного чтения при уровне изоляции READ UNCOMMITTED,
// а также её устранение использованием READ COMMITTED.
// NOTE: выяснилось, что PostgreSQL не поддерживает "грязные чтения", поэтому негативный кейс проверить невозможно.
func TestDirtyRead(t *testing.T) {
	ctx := context.Background()
	conn := connect(ctx, t)
	defer conn.Close()

	testFlow := func(readCommitted bool) {
		// Создаём тестовую таблицу: ID пользователя и его баланс.
		tableName := utils.RandomString(32)
		_, err := conn.Exec(
			ctx,
			fmt.Sprintf(
				`
			CREATE TABLE %v
			(
				id TEXT PRIMARY KEY,
				balance INTEGER
			);
			`,
				tableName,
			),
		)
		defer dropTable(ctx, t, conn, tableName)
		require.NoError(t, err)

		// Сохраняем исходные данные в таблице: у пользователя "Боб" баланс 300 руб.
		_, err = conn.Exec(ctx, fmt.Sprintf(`INSERT INTO %v VALUES ($1, $2);`, tableName), "bob", 300)
		require.NoError(t, err)

		// Начинаем транзакцию №1.
		// В зависимости от сценария проставляем уровень изоляции транзакции №1: READ COMMITTED или READ UNCOMMITTED.
		opts := pgx.TxOptions{IsoLevel: pgx.ReadUncommitted}
		if readCommitted {
			opts.IsoLevel = pgx.ReadCommitted
		}
		tx1, err := conn.BeginTx(ctx, opts)
		require.NoError(t, err)

		// Начинаем транзакцию №2.
		// Уровень изоляции транзакции №2 есть READ COMMITTED, т.е. по умолчанию.
		tx2, err := conn.Begin(ctx)
		require.NoError(t, err)
		// Считываем баланс Боба (300 руб.) и отнимаем от него 100 руб.
		// Итого в рамках транзакции №2 у Боба осталось 200 руб. Не завершаем транзакцию №2.
		balanceTx2 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&balanceTx2))
		_, err = tx2.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), balanceTx2-100, "bob")
		require.NoError(t, err)

		// Считываем баланс Боба в рамках транзакции №1.
		balanceTx1 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&balanceTx1))
		if readCommitted {
			// При уровне изоляции READ COMMITTED изменений баланса не наблюдаем, так как транзакция №2 не завершена.
			// Здесь баланс Боба по-прежнему составляет 300 руб.
			assert.Equal(t, 300, balanceTx1)
		} else {
			// При уровне изоляции READ UNCOMMITTED видим изменения баланса, внесённые ещё не завершённой транзакцией №2.
			// Здесь баланс Боба составляет 200 руб.
			// NOTE: PostgreSQL не поддерживает "грязного чтения", кейс проверить невозможно.
			// assert.Equal(t, 300-100, balanceTx1)
		}

		// Отменяем транзакцию №2.
		require.NoError(t, tx2.Rollback(ctx))

		// В рамках транзакции №1 пополняем баланс Боба на 500 руб., используя считанное значение баланса.
		_, err = tx1.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), balanceTx1+500, "bob")
		assert.NoError(t, err)

		// Завершаем транзакцию №1 .
		require.NoError(t, tx1.Commit(ctx))

		// Проверяем баланс Боба.
		actualBalance := 0
		require.NoError(t, conn.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&actualBalance))
		if readCommitted {
			// При уровне изоляции READ COMMITTED на транзакцию №1 не повлияла отменённая транзакции №2.
			// Баланс Боба составляет 800 руб, каким и должен быть.
			assert.Equal(t, 300+500, actualBalance)
		} else {
			// При уровне изоляции READ UNCOMMITTED в транзакцию №1 попали действия из отменённой транзакции №2.
			// Поэтому баланс Боба на 100 руб. меньше, чем должен быть, т.е. 700 руб.
			// NOTE: PostgreSQL не поддерживает "грязного чтения", кейс проверить невозможно.
			// assert.Equal(t, 300+500-100, actualBalance)
		}
	}

	t.Run("anomaly on READ UNCOMMITTED", func(t *testing.T) { testFlow(false) })
	t.Run("proper behavior on READ COMMITTED", func(t *testing.T) { testFlow(true) })
}

func dropTable(ctx context.Context, t *testing.T, conn *pgxpool.Pool, tableName string) {
	_, err := conn.Exec(ctx, fmt.Sprintf("DROP TABLE %v", tableName))
	assert.NoError(t, err)
}
