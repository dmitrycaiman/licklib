package blindsaga

const (
	ct = "application/json"

	ActionDo   = "DO"
	ActionUndo = "UNDO"
)

// Config есть конфигурация саги, необходимая для инициализации.
type Config struct {
	Stages    []Stage
	HostStage Stage
	SagaName  string
}

// Notification есть уведомление, доставляемое участникам саги.
type Notification struct {
	SagaName      string
	SagaID        string
	TransactionID string
	Action        string
	*Meta
}

// Meta есть хранилище информации, передающееся между этапами.
type Meta struct {
	Straight    map[string][]byte // Хранилище, заполняемое при прямом прохождении.
	Undo        map[string][]byte // Хранилище, заполняемое при обратном прохождении.
	FailedStage *FailedStage      // Информация об этапе, на котором случиласб ошибка.
}

// Failed представляет этап, завершившийся неудачей.
type FailedStage struct {
	*Stage
	Details string
}

// Stage представляет этап саги.
type Stage struct {
	Name    string
	Address string
}
