package blindsaga

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// BlindSaga является простейшим оркестратором паттерна "сага":
// позволяет синхронизировать HTTP-уведомления о необходимости выполнения или отката определённого действия.
// Позволяет организовать т.н. "слепую сагу", то есть распределённую транзакцию, успех этапов которой определяется косвенно.
type BlindSaga struct {
	stages    []*Stage     // Этапы саги.
	client    *http.Client //
	hostStage *Stage       // Информация об оркестраторе.
	sagaName  string       // Имя саги, переданное извне.
	sagaID    string       // Уникальный идентификатор саги.
}

// New создаёт новую сагу на основе конфигурации. Если HTTP-клиент равен nil, то сага будет использовать http.DefaultClient.
func New(config *Config, httpClient *http.Client) (*BlindSaga, error) {
	if len(config.Stages) == 0 {
		return nil, fmt.Errorf("empty saga")
	}
	newBlindSaga := &BlindSaga{
		hostStage: &Stage{config.HostStage.Name, config.HostStage.Address},
		sagaName:  config.SagaName,
		sagaID:    uuid.NewString(),
		client:    httpClient,
	}
	if httpClient == nil {
		newBlindSaga.client = http.DefaultClient
	}
	uniqueCheck := map[string]struct{}{}
	for _, v := range config.Stages {
		if _, ok := uniqueCheck[v.Name]; ok {
			return nil, fmt.Errorf("non-unique stage name %v", v.Name)
		}
		uniqueCheck[v.Name] = struct{}{}
		newBlindSaga.stages = append(newBlindSaga.stages, &Stage{v.Name, v.Address})
	}
	return newBlindSaga, nil
}

// Start запускает сагу с некоторой начальной информацией, возвращает накопленную метаинформацию и флаг успеха.
// Одновременно может быть запущено несколько саг.
func (slf *BlindSaga) Start(meta []byte) (*Meta, bool) {
	// Формируем уведомление о необходимости выполнить действие.
	notification := &Notification{
		SagaName:      slf.sagaName,
		SagaID:        slf.sagaID,
		TransactionID: uuid.NewString(), // Каждому эксземпляру саги присваивается идентификатор.
		Action:        ActionDo,
		Meta:          &Meta{Straight: map[string][]byte{slf.hostStage.Name: meta}},
	}
	failedStage := -1
	// Проход по этапам саги в прямом направлении.
	for i, stage := range slf.stages {
		meta, err := send(stage, notification)
		if len(meta) != 0 {
			notification.Meta.Straight[stage.Name] = meta
		}
		// При возникновении ошибки останавливаем сагу.
		if err != nil {
			notification.Action = ActionUndo
			notification.FailedStage = &FailedStage{Stage: stage, Details: err.Error()}
			failedStage = i
			break
		}
	}
	// При неудаче на определённом этапе саги уведомляем предыдущие этапы об отмене действия в обратном порядке.
	if failedStage != 0 {
		notification.Meta.Undo = map[string][]byte{}
		for i := failedStage - 1; i >= 0; i-- {
			meta, err := send(slf.stages[i], notification)
			if err != nil {
				meta = []byte(err.Error())
			}
			notification.Meta.Undo[slf.stages[i].Name] = meta
		}
	}
	return notification.Meta, failedStage == -1
}

// send осуществляет отправку уведомления по HTTP.
func send(stage *Stage, notification *Notification) ([]byte, error) {
	body, err := json.Marshal(notification)
	if err != nil {
		return nil, err
	}
	res, err := http.Post(stage.Address, ct, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %v", res.StatusCode)
	}
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
