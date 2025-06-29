package linkname

import (
	"fmt"
	"licklib/linkname/unexported"
	"math/rand"
	"testing"
	_ "unsafe" // Обязательно импортировать.

	"github.com/stretchr/testify/assert"
)

// Линк неэкспортируемой функции.
//
//go:linkname alienFunc licklib/linkname/unexported.privateFunc
func alienFunc() int

// Линк неэкспортируемой функции-переменной.
//
//go:linkname alienFuncVar licklib/linkname/unexported.privateFuncVar
var alienFuncVar func() int

// Линк неэкспортируемых переменных.
//
//go:linkname alienIntVar licklib/linkname/unexported.privateIntVar
var alienIntVar int

//go:linkname alienStringVar licklib/linkname/unexported.privateStringVar
var alienStringVar string

// Линк неэкспортируемой переменной неэкспортируемого типа.
//
//go:linkname alienPrivateTypeVar licklib/linkname/unexported.privateTypeVar
var alienPrivateTypeVar *struct{ a, b int }

// Линк неэкспортируемой переменной экспортируемого типа.
//
//go:linkname alienPublicTypeVar licklib/linkname/unexported.pubilcTypeVar
var alienPublicTypeVar *unexported.PublicType

func TestLinkname(t *testing.T) {
	// Использование неэкспортируемой функции.
	assert.Equal(t, unexported.PrivateFunc(), alienFunc())

	// Подмена реализации неэкспортирумой функции-переменной.
	assert.Equal(t, unexported.PrivateFuncVar(), alienFuncVar())
	someValue := rand.Int()
	alienFuncVar = func() int { return someValue }
	assert.Equal(t, someValue, unexported.PrivateFuncVar())

	// Использование неэкспортируемых переменных.
	assert.Equal(t, unexported.PrivateIntVar(), alienIntVar)
	assert.Equal(t, unexported.PrivateStringVar(), alienStringVar)
	// Изменение неэкспортируемых переменных.
	alienIntVar, alienStringVar = rand.Int(), fmt.Sprint(rand.Int())
	assert.Equal(t, unexported.PrivateIntVar(), alienIntVar)
	assert.Equal(t, unexported.PrivateStringVar(), alienStringVar)

	// Использование неэкспортируемой переменной неэкспортируемого типа.
	assert.Equal(t, unexported.PrivateTypeVarA(), alienPrivateTypeVar.a)
	assert.Equal(t, unexported.PrivateTypeVarB(), alienPrivateTypeVar.b)
	// Изменение неэкспортируемой переменной неэкспортируемого типа.
	alienPrivateTypeVar.a, alienPrivateTypeVar.b = rand.Int(), rand.Int()
	assert.Equal(t, unexported.PrivateTypeVarA(), alienPrivateTypeVar.a)
	assert.Equal(t, unexported.PrivateTypeVarB(), alienPrivateTypeVar.b)

	// Использование неэкспортируемой переменной экспортируемого типа.
	assert.Equal(t, unexported.PubilcTypeVar(), alienPublicTypeVar)
	// Изменение неэкспортируемой переменной экспортируемого типа.
	alienPublicTypeVar = &unexported.PublicType{A: rand.Int(), B: rand.Int()}
	assert.Equal(t, unexported.PubilcTypeVar(), alienPublicTypeVar)
}
