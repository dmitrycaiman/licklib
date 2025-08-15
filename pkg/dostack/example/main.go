package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"licklib/pkg/dostack"
	"log"
	"os"
	"runtime"
	"strings"
)

type collector struct {
	memStats *runtime.MemStats
}

func (slf *collector) collect() { runtime.ReadMemStats(slf.memStats) }

func (slf *collector) flush(w io.Writer) error {
	b, err := json.MarshalIndent(slf.memStats, "", " ")
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	slf.memStats = &runtime.MemStats{}
	return nil
}

func main() {
	c := &collector{&runtime.MemStats{}}

	ds := dostack.New(
		dostack.WithDoer(
			"save",
			&cmdSaveMemstats{c, "metrics.json"},
			false,
		),
		dostack.WithDoer(
			"reset",
			&cmdResetMemstats{c: c},
			true,
		),
		dostack.WithDoer(
			"print",
			&cmdPrintMemstats{c: c},
			false,
		),
		dostack.WithDoer(
			"update",
			&cmdUpdateMemstats{c: c},
			false,
		),
		dostack.WithExplicitUndo("undo"),
	)

	scn := bufio.NewScanner(os.Stdin)
	for {
		if !scn.Scan() {
			log.Fatal(scn.Err())
		}

		t := scn.Text()
		if strings.HasPrefix(t, "exit") {
			return
		}
		if err := ds.Do(t); err != nil {
			log.Printf("command invoke error: %v\n", err)
		}
		fmt.Println()
	}
}
