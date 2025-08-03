package meanval

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeanval(t *testing.T) {
	m := NewMeanval()
	m.Upsert("a", NewAverage())
	m.Upsert("b", NewMode())
	m.Upsert("c", NewMedian())

	var wg sync.WaitGroup
	for i := range 1000 {
		wg.Add(4)
		go func() {
			m.Update("a", i)
			wg.Done()
		}()
		go func() {
			m.Update("b", i)
			wg.Done()
		}()
		go func() {
			m.Update("c", i)
			wg.Done()
		}()
		go func() {
			assert.Equal(t, NonExistentValue, m.Update(fmt.Sprint(i), i))
			wg.Done()
		}()
	}
	wg.Wait()

	assert.Equal(t, 499, m.Select("a"))
	assert.GreaterOrEqual(t, 1000, m.Select("b"))
	assert.Equal(t, 500, m.Select("c"))

	m.Reset("a")
	assert.NotZero(t, m.points["a"])
	assert.Zero(t, m.storage["a"])
	assert.Equal(t, NonExistentValue, m.Select("a"))

	m.Reset("")
	assert.Empty(t, m.storage)

	m.Delete("b")
	assert.Zero(t, m.points["b"])
	assert.Zero(t, m.storage["b"])

	m.Delete("")
	assert.Empty(t, m.points)
	assert.Empty(t, m.storage)
}
