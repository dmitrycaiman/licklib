package main

import "runtime"

type cmdResetMemstats struct {
	c    *collector
	prev *runtime.MemStats
}

func (slf *cmdResetMemstats) Do() error {
	slf.prev = slf.c.memStats
	slf.c.memStats = &runtime.MemStats{}
	return nil
}

func (slf *cmdResetMemstats) Undo() error {
	slf.c.memStats = slf.prev
	return nil
}
