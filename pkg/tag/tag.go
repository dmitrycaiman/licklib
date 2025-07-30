package tag

import (
	"fmt"
	"log"
	"strings"
)

type LogFunc func(string, ...any)

type Tag struct {
	logFunc        LogFunc
	tag            string
	defaultLogFunc bool
}

func New(tags ...string) *Tag {
	return &Tag{
		logFunc:        log.Printf,
		defaultLogFunc: true,
		tag:            "[" + strings.Join(tags, "][") + "] ",
	}
}

func (slf Tag) T(input string) string { return fmt.Sprint(slf.tag, input) }

func (slf *Tag) Log(template string, args ...any) {
	if slf.defaultLogFunc {
		template += "\n"
	}
	slf.logFunc(slf.T(template), args...)
}

func (slf *Tag) Errorf(template string, args ...any) error {
	return fmt.Errorf(slf.T(template), args...)
}

func (slf *Tag) Error(err error) error { return fmt.Errorf("%v%w", slf.tag, err) }

func (slf *Tag) WithLogFunc(logFunc LogFunc) *Tag {
	if logFunc != nil {
		slf.logFunc = logFunc
		slf.defaultLogFunc = false
	}
	return slf
}
