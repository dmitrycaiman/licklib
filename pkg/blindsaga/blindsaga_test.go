package blindsaga

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const hostName = "host"

type testServer struct {
	t        *testing.T
	server   *httptest.Server
	_ID      string
	hostCode int
	willFail bool
	counter  *struct{ n int }
}

func (slf *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Эмулируем ошибку на этапе.
	if slf.willFail {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	assert.NoError(slf.t, err)
	n := &Notification{}
	assert.NoError(slf.t, json.Unmarshal(body, &n))

	switch n.Action {
	case ActionDo:
		// Считываем код хоста и запоминаем.
		code, err := strconv.Atoi(string(n.Straight[hostName]))
		assert.NoError(slf.t, err)
		slf.hostCode = code

		// В качестве ответа отправляем текущее значение счетчика и инкрементируем его.
		w.Write(fmt.Append(nil, slf.counter.n))
		slf.counter.n++
	case ActionUndo:
		// В качестве ответа отправляем текущее значение счетчика и декрементируем его.
		w.Write(fmt.Append(nil, slf.counter.n))
		slf.counter.n--
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func TestBlindSaga(t *testing.T) {
	testFlow := func(failure bool) {
		config := &Config{HostStage: Stage{Name: hostName}, SagaName: uuid.NewString()}
		hostCode := rand.Int() // Код хоста. Идея в том, чтобы переслать его всем участникам саги.
		expectedMeta := map[string][]byte{config.HostStage.Name: fmt.Append(nil, hostCode)}
		counter := &struct{ n int }{}

		servers := []*testServer{}
		for i := range 10 {
			server := &testServer{t: t, _ID: uuid.NewString(), counter: counter}
			server.server = httptest.NewServer(server)
			servers = append(servers, server)
			expectedMeta[servers[i]._ID] = fmt.Append(nil, i)                           // Каждый сервер должен прислать свой порядковый номер.
			config.Stages = append(config.Stages, Stage{server._ID, server.server.URL}) // Добавляем сервер в конфиг в порядке возрастания.
			if failure && i == 9 {
				server.willFail = true
			}
		}

		// Создаём сагу.
		s, err := New(config, nil)
		assert.NoError(t, err)
		assert.Len(t, s.stages, len(servers))

		// Запускаем сагу.
		actualMeta, ok := s.Start(fmt.Append(nil, hostCode))
		if failure {
			assert.False(t, ok)
		} else {
			assert.True(t, ok)
		}

		// Проверяем, что каждый этап выслал свой порядковый номер, что скает о верном порядке выполнения этапов.
		for k, actualData := range actualMeta.Straight {
			expectedData, ok := expectedMeta[k]
			assert.True(t, ok)
			assert.Equal(t, expectedData, actualData)
			delete(expectedMeta, k)
		}
		if failure {
			// При отмене действий каждый сервер декрементировал счетчик.
			assert.Zero(t, counter.n)
			// Последний сервер выдал ошибку.
			assert.Equal(t, servers[9]._ID, actualMeta.FailedStage.Name)
			assert.Equal(t, servers[9].server.URL, actualMeta.FailedStage.Address)
			assert.Len(t, expectedMeta, 1)
			// Убеждаемся, что на каждом этапе, кроме последнего, был прочитан код хоста.
			for i := 0; i < len(servers)-1; i++ {
				assert.Equal(t, hostCode, servers[i].hostCode)
			}
		} else {
			// Каждый сервер инкрементировал счетчик.
			assert.Equal(t, 10, counter.n)
			assert.Empty(t, expectedMeta)
			// Убеждаемся, что на каждом этапе был прочитан код хоста.
			for _, v := range servers {
				assert.Equal(t, hostCode, v.hostCode)
			}
		}

	}

	t.Run("success", func(t *testing.T) { testFlow(false) })
	t.Run("failure", func(t *testing.T) { testFlow(true) })
}
