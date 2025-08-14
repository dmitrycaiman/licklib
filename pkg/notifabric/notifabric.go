package notifabric

import (
	"fmt"
	"io"
)

const (
	File = "file"
	Http = "http"
	Log  = "log"
)

type Notifier interface{ Notify(message string) error }

// Notifabric есть фабрика по созданию уведомителей.
type Notifabric struct {
	outputFile      string
	httpDestination string
	logTag          string
	customWriters   map[string]io.Writer
}

// New создаёт фабрику уведомителей. По типу уведомителя будет выдана конкретная настроенная реализация.
// Имеется возможность также передать пользовательские интерфейсы io.Writer как способы отправки уведомлений.
func New(outputFile, httpDestination, ioStream string, custom map[string]io.Writer) *Notifabric {
	return &Notifabric{outputFile, httpDestination, ioStream, custom}
}

func (slf *Notifabric) CreateNotificator(kind string) (Notifier, error) {
	switch kind {
	case File:
		return &fileNotifier{slf.outputFile}, nil
	case Http:
		return &httpNotifier{slf.httpDestination}, nil
	case Log:
		return &logNotifier{slf.logTag}, nil
	default:
		if slf.customWriters != nil {
			if w, ok := slf.customWriters[string(kind)]; ok {
				return &customNotifier{w}, nil
			}
		}
	}
	return nil, fmt.Errorf("unknown notification type")
}
