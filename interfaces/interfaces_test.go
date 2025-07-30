package interfaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type I interface{ F() }

type S struct{}

func (slf *S) F() {}

func TestNilInterface(t *testing.T) {
	// Созданная переменная интерфейсного типа является nil, так как интерфейс есть указатель.
	var i I
	assert.True(t, i == nil)
	// Интерфейс, в который помещен nil-указатель, уже не является nil, так как есть информация о типе.
	var s *S
	i = s
	assert.True(t, s == nil)
	assert.False(t, i == nil)
}
