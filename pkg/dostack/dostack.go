package dostack

import (
	"fmt"
	"sync"
	"time"
)

// loggedAction есть элемент журнала совершённых действий и откатов.
type loggedAction struct {
	name      string    // Имя действия.
	timestamp time.Time // Время совершения.
	undo      bool      // Флаг того, что действиие было откатом.
}

// command представляет единицу исполнения.
type command struct {
	Doer          // Переданный извне исполнитель.
	undoable bool // Флаг того, поддерживает ли исполнитель откат.
}

// Dostack есть хранилище команд. Выполненные команды кладутся в стек.
// В порядке извлечения из стека можно производить откат.
type Dostack struct {
	mu       sync.Mutex
	commands map[string]*command // Список зарегистрированных команд.
	stack    []string            // Стек имён совершённых действий.
	log      []loggedAction      // Журнал совершённых действий и откатов.
}

// New создаёт новое хранилище.
func New(options ...Option) *Dostack {
	dostack := &Dostack{commands: map[string]*command{}}
	for _, v := range options {
		v(dostack)
	}
	return dostack
}

// Do исполняет зарегистрированную команду.
func (slf *Dostack) Do(name string) error {
	slf.mu.Lock()
	defer slf.mu.Unlock()

	command, ok := slf.commands[name]
	if !ok {
		return fmt.Errorf("command <%v> is not registered", name)
	}

	if err := safeExec(command.Do); err != nil {
		return fmt.Errorf("command <%v> execution failure: %w", name, err)
	}

	slf.log = append(slf.log, loggedAction{name: name, timestamp: time.Now().UTC()})
	slf.stack = append(slf.stack, name)

	return nil
}

// Undo вызывает откат последнего совершённого действия, если он был поддержан.
func (slf *Dostack) Undo() error {
	slf.mu.Lock()
	defer slf.mu.Unlock()

	if len(slf.stack) == 0 {
		return nil
	}

	lastActionName := slf.stack[len(slf.stack)-1]
	command := slf.commands[lastActionName]
	if command.undoable {
		if err := safeExec(command.Undo); err != nil {
			return err
		}
	}

	slf.log = append(slf.log, loggedAction{name: lastActionName, timestamp: time.Now().UTC(), undo: true})
	slf.stack = slf.stack[:len(slf.stack)-1]

	return nil
}

// AddDoer добавляет к списку действий сущность, реализующую интерфейс команды Doer.
// Флаг undoable есть признак того, что команда поддерживает откат.
func (slf *Dostack) AddDoer(name string, doer Doer, undoable bool) {
	WithDoer(name, doer, undoable)(slf)
}

// AddFunc добавляет к списку действий функцию. Запрос отката действия невозможен.
func (slf *Dostack) AddFunc(name string, do func() error) {
	WithFunc(name, do)(slf)
}

// AddFuncs добавляет к списку действий функции непосредственно действия и отката.
func (slf *Dostack) AddFuncs(name string, do, undo func() error) {
	WithFuncs(name, do, undo)(slf)
}

// safeExec восстанавливает исполнение из состояния паники при необходимости.
func safeExec(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()
	return f()
}
