package reflect

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type I interface {
	F() int
	f() testInt
}

type testInt int

type S struct {
	i    int     `one:"tag-for-i"`
	j    testInt `two:"tag-for-j"`
	K    float64 `three:"tag-for-K"`
	uint `four:"tag-for-uint"`
}

func (slf *S) F() int     { return slf.i }
func (slf *S) f() testInt { return slf.j }
func (slf S) Z() float64  { return slf.K }

func TestReflectType(t *testing.T) {
	// Создаём переменную типа интерфейс. По сути нулевой указатель.
	var i I

	// На nil-интерфейс TypeOf возвращает nil.
	rt := reflect.TypeOf(i)
	assert.Nil(t, rt)

	// Присваиваем интерфейсу значение в виде нулевого указателя на тип.
	var s *S = &S{12, 34, 56, 78}
	i = s

	// Теперь TypeOf возвращает информацию о типе.
	rt = reflect.TypeOf(i)
	assert.NotNil(t, rt)
	assert.Equal(t, reflect.Pointer, rt.Kind())                     // Под интерфейсом скрывался указатель.
	assert.Empty(t, rt.Name())                                      // Имя типа пустое, т.к. это указатель.
	assert.Panics(t, func() { rt.NumField() })                      // Паникует при взятии числа полей и при других операциях с полями.
	assert.Equal(t, 2, rt.NumMethod())                              // Есть информация только об экспортируемых методах, т.к. работаем не с интерфейсом.
	assert.True(t, rt.Implements(reflect.TypeOf((*I)(nil)).Elem())) // Проверка на то, что мы реализуем I.

	// У типа "указатель" можно исследовать все экспортируемые методы.
	assert.Panics(t, func() { rt.Method(2) }) // Паника при взятии несуществующего метода по номеру.
	rtm := rt.Method(0)
	assert.NotNil(t, rtm)
	assert.Equal(t, 0, rtm.Index)                  // Проверка индекса.
	assert.Equal(t, 1, rt.Method(1).Index)         //
	assert.Equal(t, "F", rtm.Name)                 // Проверка имени.
	assert.Equal(t, "Z", rt.Method(1).Name)        //
	assert.Equal(t, reflect.Func, rtm.Type.Kind()) // Тип метода есть "функция".
	assert.True(t, rtm.IsExported())               // Метод экспортируемый.

	// Можно исследовать функцию экспортируемого метода, которая является reflect.Value.
	assert.False(t, rtm.Func.IsNil())
	assert.Equal(t, reflect.Func, rtm.Func.Kind())                                         // Проверка типа "функция".
	assert.True(t, rtm.Func.IsValid())                                                     // Проверка на валидность проходит.
	assert.False(t, rtm.Func.CanSet())                                                     // Функцию нельзя переприсвоить.
	assert.Equal(t, 12, rtm.Func.Call([]reflect.Value{reflect.ValueOf(s)})[0].Interface()) // При передаче ресивера её можно запустить.
	assert.Equal(t, -12, rtm.Func.Call([]reflect.Value{reflect.ValueOf(&S{i: -12})})[0].Interface())

	// Можно извлечь информацию о низлежащем типе, если полученный тип есть указатель.
	rte := rt.Elem()
	assert.Panics(t, func() { rte.Elem() })                           // Паника при повторном разыменовании.
	assert.Equal(t, reflect.Struct, rte.Kind())                       // Извлекли тип общий тип "структура".
	assert.Equal(t, "S", rte.Name())                                  // Можем прочитать имя типа.
	assert.False(t, rte.Implements(reflect.TypeOf((*I)(nil)).Elem())) // Уже не реализуем I, так как являемся структурой.
	assert.Equal(t, 1, rte.NumMethod())                               // Есть информация экспортируемых методах c non-pointer receiver.

	// У типа "структура" можно исследовать только методы c non-pointer receiver.
	rtem := rte.Method(0)
	assert.Equal(t, 0, rtem.Index)                                                                    // Проверка индекса.
	assert.Equal(t, "Z", rtem.Name)                                                                   // Проверка имени.
	assert.Equal(t, reflect.Func, rtm.Type.Kind())                                                    // Тип метода есть "функция".
	assert.True(t, rtm.IsExported())                                                                  // Метод экспортируемый.
	assert.Equal(t, float64(56), rtem.Func.Call([]reflect.Value{reflect.ValueOf(*s)})[0].Interface()) // При передаче non-pointer receiver функцию метода можно запустить.

	// У типа "структура" можно исследовать поля.
	assert.Equal(t, 4, rte.NumField())        // Информация об всех полях (экспортируемых и нет).
	assert.Panics(t, func() { rte.Field(4) }) // Паника при взятии несуществующего поля по номеру.
	// Извлечение полей.
	f1, f2, f3, f4 := rte.Field(0), rte.Field(1), rte.Field(2), rte.Field(3)
	assert.Equal(t, "i", f1.Name)                       // Проверка имён.
	assert.Equal(t, "j", f2.Name)                       //
	assert.Equal(t, "K", f3.Name)                       //
	assert.Equal(t, "uint", f4.Name)                    //
	assert.Equal(t, "tag-for-i", f1.Tag.Get("one"))     // Проверка структурных тегов.
	assert.Equal(t, "tag-for-j", f2.Tag.Get("two"))     //
	assert.Equal(t, "tag-for-K", f3.Tag.Get("three"))   //
	assert.Equal(t, "tag-for-uint", f4.Tag.Get("four")) //
	assert.False(t, f1.IsExported())                    // Проверка экспортируемости.
	assert.False(t, f2.IsExported())                    //
	assert.True(t, f3.IsExported())                     //
	assert.False(t, f4.IsExported())                    //
	assert.False(t, f1.Anonymous)                       // Проверка анонимности.
	assert.False(t, f2.Anonymous)                       //
	assert.False(t, f3.Anonymous)                       //
	assert.True(t, f4.Anonymous)                        //
	assert.Equal(t, reflect.Int, f1.Type.Kind())        // Проверка типа поля.
	assert.Equal(t, reflect.Int, f2.Type.Kind())        // Пользовательский testInt всё равно является int.
	assert.Equal(t, reflect.Float64, f3.Type.Kind())    //
	assert.Equal(t, reflect.Uint, f4.Type.Kind())       //

	// Также рассмотрим указатель на интерфейс.
	var pi *I = &i

	// Исследуем тип.
	rt = reflect.TypeOf(pi)
	assert.NotNil(t, rt)
	assert.Equal(t, reflect.Pointer, rt.Kind()) // Тип есть указатель.
	assert.Empty(t, rt.Name())                  // Имя типа пустое, т.к. это указатель.
	assert.Panics(t, func() { rt.NumField() })  // Паникует при взятии числа полей и при других операциях с полями.
	assert.Zero(t, rt.NumMethod())              // Информация о методах отсутствует.

	// Разыменуем указатель.
	rte = rt.Elem()
	assert.Panics(t, func() { rte.Elem() })        // Паника при повторном разыменовании.
	assert.Equal(t, reflect.Interface, rte.Kind()) // Извлекли тип общий тип "интерфейс".
	assert.Equal(t, "I", rte.Name())               // Можем прочитать имя типа.
	assert.Equal(t, 2, rte.NumMethod())            // Есть информация об экспортируемых и неэкспортируемых методах.

	// Исследуем методы.
	rtem1, rtem2 := rte.Method(0), rte.Method(1)
	assert.Equal(t, "F", rtem1.Name)
	assert.Equal(t, "f", rtem2.Name)
	assert.False(t, rtem1.Func.IsValid()) // Проверка на валидность.
	assert.False(t, rtem2.Func.IsValid())
	assert.True(t, rtem1.IsExported())  // Проверка экспортируемости.
	assert.False(t, rtem2.IsExported()) //
}

func TestReflectValue(t *testing.T) {
	// Создаём переменную типа интерфейс. По сути нулевой указатель.
	var i I

	// На nil-интерфейс ValueOf возвращает Invalid Value.
	// Данное значение приводит к панике в большинстве случаев использования.
	rv := reflect.ValueOf(i)
	assert.Empty(t, rv)                         // Возвращается по сути reflect.Value{}.
	assert.Equal(t, reflect.Invalid, rv.Kind()) // Под Invalid Value есть конкретный Kind.
	assert.False(t, rv.IsValid())               // Явная проверка на невалидность.
	assert.Panics(t, func() { rv.IsNil() })     // Паника даже при проверке на nil.

	// Присваиваем интерфейсу значение в виде нулевого указателя на тип.
	var s *S = &S{12, 34, 56, 78}
	i = s

	// Теперь ValueOf возвращает информацию о значени.
	rv = reflect.ValueOf(i)
	assert.NotNil(t, rv)
	assert.Equal(t, reflect.Pointer, rv.Kind()) // Под интерфейсом скрывался указатель.
	assert.False(t, rv.CanAddr())               //
	assert.Panics(t, func() { rv.NumField() })  // Паникует при взятии числа полей и при других операциях с полями.
	assert.Equal(t, 2, rv.NumMethod())          // Есть информация только об экспортируемых методах, т.к. работаем не с интерфейсом.

	// Исследуем все экспортируемые методы.
	assert.Panics(t, func() { rv.Method(2) })                           // Паника при взятии несуществующего метода по номеру.
	assert.Equal(t, reflect.Func, rv.Method(0).Kind())                  // Тип метода есть "функция".
	assert.Equal(t, 12, rv.Method(0).Call(nil)[0].Interface())          // Функцию можно запустить, ресивер передавать не нужно.
	assert.Equal(t, float64(56), rv.Method(1).Call(nil)[0].Interface()) //

	// Можно извлечь информацию о лежащем под указателем значении.
	rve := rv.Elem()
	assert.Panics(t, func() { rve.Elem() })     // Паника при повторном разыменовании.
	assert.Equal(t, reflect.Struct, rve.Kind()) // Извлекли тип общий тип "структура".
	assert.Equal(t, 1, rve.NumMethod())         // Есть информация экспортируемых методах c non-pointer receiver.

	// У типа "структура" можно исследовать только методы c non-pointer receiver.
	rvem := rve.Method(0)
	assert.Equal(t, reflect.Func, rvem.Kind())                  // Тип метода есть "функция".
	assert.Equal(t, float64(56), rvem.Call(nil)[0].Interface()) // Функцию метода можно запустить, ресивер передавать не нужно.

	// У типа "структура" можно исследовать поля.
	assert.Equal(t, 4, rve.NumField())        // Информация об всех полях (экспортируемых и нет).
	assert.Panics(t, func() { rve.Field(4) }) // Паника при взятии несуществующего поля по номеру.
	// Извлечение полей.
	f1, f2, f3, f4 := rve.Field(0), rve.Field(1), rve.Field(2), rve.Field(3)
	assert.Panics(t, func() { f1.Interface() })  // Взятие значения неэкспортируемых полей ведёт к панике.
	assert.Panics(t, func() { f2.Interface() })  //
	assert.Panics(t, func() { f4.Interface() })  //
	assert.Equal(t, float64(56), f3.Interface()) // Можно взять значение экспортируемого поля.

	// Также рассмотрим указатель на интерфейс.
	var pi *I = &i

	// Исследуем значение.
	rv = reflect.ValueOf(pi)
	assert.NotNil(t, rv)
	assert.Equal(t, reflect.Pointer, rv.Kind()) // Тип есть указатель.
	assert.Panics(t, func() { rv.NumField() })  // Паникует при взятии числа полей и при других операциях с полями.
	assert.Zero(t, rv.NumMethod())              // Информация о методах отсутствует.

	// Разыменуем указатель.
	rve = rv.Elem()
	assert.Equal(t, reflect.Interface, rve.Kind())              // Извлекли тип общий тип "интерфейс".
	assert.Equal(t, 2, rve.NumMethod())                         // Есть информация об экспортируемых и неэкспортируемых методах.
	assert.Equal(t, 12, rve.Method(0).Call(nil)[0].Interface()) // Экспортируемую функцию интерфейса можно запустить.
	assert.Panics(t, func() { rve.Method(1).Call(nil) })        // Вызов неэкспортируемой функции интерфейса приведёт к панике.
}
