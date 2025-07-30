package dostack

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testDoer struct{ i int }

func (slf *testDoer) Do() error {
	slf.i++
	return nil
}

func (slf *testDoer) Undo() error {
	slf.i--
	return nil
}

func TestDostack(t *testing.T) {
	var i int
	d := &testDoer{}
	testDostack := New(
		WithFuncs(
			"a",
			func() error {
				i++
				return nil
			},
			func() error {
				i--
				return nil
			},
		),
		WithFunc(
			"b",
			func() error {
				i++
				return nil
			},
		),
		WithDoer("c", d, true),
	)

	var wg sync.WaitGroup
	for range 1_000 {
		wg.Add(3)
		go func() {
			defer wg.Done()
			testDostack.Do("a")
		}()
		go func() {
			defer wg.Done()
			testDostack.Do("b")
		}()
		go func() {
			defer wg.Done()
			testDostack.Do("c")
		}()
	}
	wg.Wait()
	for range 3_000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testDostack.Undo()
		}()
	}
	wg.Wait()

	assert.Equal(t, 1000, i)
	assert.Equal(t, 0, d.i)
}
