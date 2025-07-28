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

func TestNonRepeatableRead(t *testing.T) {
	ctx := context.Background()
	conn := connect(ctx, t)
	defer conn.Close()

	testFlow := func(repeatableRead bool) {
		// Создаём тестовую таблицу.
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
		_, err = conn.Exec(ctx, fmt.Sprintf(`INSERT INTO %v VALUES ($1, $2);`, tableName), "bob", 200)
		require.NoError(t, err)

		// Проверяем факт сохранения данных.
		actualBalance := 0
		require.NoError(t, conn.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&actualBalance))
		require.Equal(t, 200, actualBalance)

		// Начинаем транзакцию №1.
		opts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
		if repeatableRead {
			// В зависимости от кейса проставляем уровень изоляции транзакции №1.
			opts.IsoLevel = pgx.RepeatableRead
		}
		tx1, err := conn.BeginTx(ctx, opts)
		require.NoError(t, err)
		// Считываем баланс Боба.
		initBalanceTx1 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&initBalanceTx1))
		// Рассчитаем доход по вкладу на основе считанного значения.
		income := initBalanceTx1 / 10

		// Начинаем транзакцию №2. К балансу Боба добавляем 100.
		// Уровень изоляции транзакции №2 есть READ COMMITTED, т.е. по умолчанию.
		tx2, err := conn.Begin(ctx)
		require.NoError(t, err)
		balanceTx2 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&balanceTx2))
		_, err = tx2.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), balanceTx2+100, "bob")
		require.NoError(t, err)
		// Завершаем транзакцию №2.
		require.NoError(t, tx2.Commit(ctx))

		// Вновь считываем баланс Боба в транзакции №1.
		lastBalanceTx1 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&lastBalanceTx1))

		// В рамках транзакции №1 пополняем баланс Боба на сумму дохода по вкладу, рассчитанную в начале.
		_, err = tx1.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), lastBalanceTx1+income, "bob")
		if repeatableRead {
			// При уровне изоляции REPEATABLE READ произойдет ошибка сериализации, так как данные в строке были изменены в транзакции №2.
			assert.Error(t, err)
		} else {
			// При уровне изоляции READ COMMITTED ошибки не произойдёт, и к балансу Боба будет прибавлен доход, рассчитанный по устаревшим данным.
			assert.NoError(t, err)
		}
		// Завершаем транзакцию №1 в любом случае. При ошибке на предыдущем шаге можно было сделать также Rollback.
		err = tx1.Commit(ctx)
		if repeatableRead {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)

		// Вновь считываем баланс Боба после всех транзакций.
		actualBalance = 0
		require.NoError(t, conn.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&actualBalance))
		// Доход от вклада был посчитан не от конечной суммы (200+100)/10...
		assert.NotEqual(t, (200+100)/10+200+100, actualBalance)
		// А от изначальной 200/10.
		assert.Equal(t, 200/10+200+100, actualBalance)
		// Это объясняется тем, что на уровне изоляции READ COMMITTED произошла аномалия неповторяющегося чтения, вследствие чего
		// одинаковые SELECT-ы в рамках одной транзакции дали разныый результат: по одному был рассчитан доход, по другому произошло начисление.
	}

	t.Run("anomaly on READ COMMITTED", func(t *testing.T) { testFlow(false) })
	t.Run("error on REPEATABLE READ", func(t *testing.T) { testFlow(true) })
}

func TestLostUpdate(t *testing.T) {
	ctx := context.Background()
	conn := connect(ctx, t)
	defer conn.Close()

	testFlow := func(repeatableRead bool) {
		// Создаём тестовую таблицу.
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
		_, err = conn.Exec(ctx, fmt.Sprintf(`INSERT INTO %v VALUES ($1, $2);`, tableName), "bob", 200)
		require.NoError(t, err)

		// Проверяем факт сохранения данных.
		actualBalance := 0
		require.NoError(t, conn.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&actualBalance))
		require.Equal(t, 200, actualBalance)

		// Начинаем транзакцию №1. Считываем баланс Боба.
		opts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
		if repeatableRead {
			// В зависимости от кейса проставляем уровень изоляции транзакции №1.
			opts.IsoLevel = pgx.RepeatableRead
		}
		tx1, err := conn.BeginTx(ctx, opts)
		require.NoError(t, err)
		balanceTx1 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&balanceTx1))

		// Начинаем транзакцию №2. Считываем баланс Боба и отнимаем от него 100.
		// Уровень изоляции транзакции №2 есть READ COMMITTED, т.е. по умолчанию.
		tx2, err := conn.Begin(ctx)
		require.NoError(t, err)
		balanceTx2 := 0
		require.NoError(t, tx1.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&balanceTx2))
		_, err = tx2.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), balanceTx2-100, "bob")
		require.NoError(t, err)
		// Завершаем транзакцию №2.
		require.NoError(t, tx2.Commit(ctx))

		// В рамках транзакции №1 пополняем баланс Боба на 500, используя ранее считанное значение баланса.
		_, err = tx1.Exec(ctx, fmt.Sprintf(`UPDATE %v SET balance = $1 WHERE id = $2;`, tableName), balanceTx1+500, "bob")
		if repeatableRead {
			// При уровне изоляции REPEATABLE READ произойдет ошибка сериализации, так как данные в строке были изменены в транзакции №2.
			assert.Error(t, err)
		} else {
			// При уровне изоляции READ COMMITTED ошибки не произойдёт, и данные транзакции №2 будут потеряны.
			assert.NoError(t, err)
		}
		// Завершаем транзакцию №1 в любом случае. При ошибке на предыдущем шаге можно было сделать также Rollback.
		err = tx1.Commit(ctx)
		if repeatableRead {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)

		// Предполагаемый баланс после двух транзакций.
		require.NoError(t, conn.QueryRow(ctx, fmt.Sprintf(`SELECT balance FROM %v WHERE id = $1;`, tableName), "bob").Scan(&actualBalance))
		assert.NotEqual(t, 200-100+500, actualBalance)

		// Обновление транзакции №2 будет потеряно (аномалия), так как
		// уровень изоляции READ COMMITTED не выдал ошибку при повторном обновлении строки в рамках транзакции №1.
		// Транзакция №1 произвела обновление, основанное на "старом" чтении, произошедшем до действий транзакции №2.
		assert.Equal(t, 200+500, actualBalance)
	}

	t.Run("anomaly on READ COMMITTED", func(t *testing.T) { testFlow(false) })
	t.Run("error on REPEATABLE READ", func(t *testing.T) { testFlow(true) })
}

func dropTable(ctx context.Context, t *testing.T, conn *pgxpool.Pool, tableName string) {
	_, err := conn.Exec(ctx, fmt.Sprintf("DROP TABLE %v", tableName))
	assert.NoError(t, err)
}
