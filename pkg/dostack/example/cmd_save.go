package main

import "os"

type cmdSaveMemstats struct {
	c    *collector
	path string
}

func (slf *cmdSaveMemstats) Do() error {
	f, err := os.OpenFile(slf.path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	return slf.c.flush(f)
}

func (slf *cmdSaveMemstats) Undo() error { return nil }
