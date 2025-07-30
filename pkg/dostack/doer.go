package dostack

type Doer interface {
	Do() error
	Undo() error
}

type doer struct{ do, undo func() error }

func (slf *doer) Do() error   { return slf.do() }
func (slf *doer) Undo() error { return slf.undo() }
