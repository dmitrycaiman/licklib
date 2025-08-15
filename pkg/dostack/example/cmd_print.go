package main

import "os"

type cmdPrintMemstats struct{ c *collector }

func (slf *cmdPrintMemstats) Do() error { return slf.c.flush(os.Stdin) }

func (slf *cmdPrintMemstats) Undo() error { return nil }
