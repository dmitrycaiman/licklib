package notifabric

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type httpNotifier struct{ url string }

func (slf *httpNotifier) Notify(message string) error {
	res, err := http.Post(slf.url, "applicaton/json", bytes.NewBufferString(message))
	if err != nil {
		return err
	}
	if code := res.StatusCode; code > 299 {
		return fmt.Errorf("unexpected status code %v", code)
	}
	return nil
}

type fileNotifier struct{ path string }

func (slf *fileNotifier) Notify(message string) error {
	f, err := os.OpenFile(slf.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = f.WriteString(message + "\n")
	return err
}

type logNotifier struct{ tag string }

func (slf *logNotifier) Notify(message string) error {
	log.Printf("[%v] %v\n", slf.tag, message)
	return nil
}

type customNotifier struct{ w io.Writer }

func (slf *customNotifier) Notify(message string) error {
	_, err := slf.w.Write([]byte(message))
	return err
}
