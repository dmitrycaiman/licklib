package dostack

// Option предназначен для добавления свойств и команд в конструкторе.
type Option func(*Dostack)

// WithDoer добавляет к списку действий сущность, реализующую интерфейс команды Doer.
// Флаг undoable есть признак того, что команда поддерживает откат.
func WithDoer(name string, doer Doer, undoable bool) Option {
	return func(d *Dostack) {
		d.mu.Lock()
		defer d.mu.Unlock()
		d.commands[name] = &command{Doer: doer, undoable: undoable}
	}
}

// WithFunc добавляет к списку действий функцию.
// Запрос отката действия невозможен.
func WithFunc(name string, do func() error) Option {
	return func(d *Dostack) {
		d.mu.Lock()
		defer d.mu.Unlock()
		if do != nil {
			d.commands[name] = &command{Doer: &doer{do: do}, undoable: false}
		}
	}
}

// WithFuncs добавляет к списку действий функции непосредственно действия и отката.
func WithFuncs(name string, do, undo func() error) Option {
	return func(d *Dostack) {
		d.mu.Lock()
		defer d.mu.Unlock()
		if do != nil && undo != nil {
			d.commands[name] = &command{Doer: &doer{do, undo}, undoable: true}
		}
	}
}

// WithExplicitUndo добавляет к списку действий явное именованное применение отката.
func WithExplicitUndo(name string) Option {
	return func(d *Dostack) { d.AddFunc(name, d.explicitUndo) }
}

func (slf *Dostack) explicitUndo() error {
	slf.mu.Unlock()
	defer slf.mu.Lock()
	return slf.Undo()
}
