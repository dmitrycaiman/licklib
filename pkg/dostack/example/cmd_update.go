package main

import "runtime"

type cmdUpdateMemstats struct {
	c    *collector
	prev *runtime.MemStats
}

func (slf *cmdUpdateMemstats) Do() error {
	slf.prev = slf.c.memStats
	slf.c.memStats = &runtime.MemStats{}
	slf.c.collect()
	return nil
}

func (slf *cmdUpdateMemstats) Undo() error {
	slf.c.memStats = slf.prev
	return nil
}
