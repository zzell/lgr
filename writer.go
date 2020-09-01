package lgr

import (
	"errors"
	"os"
	"sync"
)

const (
	LF byte = 10 // Line Feed '\n'
	CR byte = 13 // Carriage Return '\r'
)

type output int

const (
	stdout output = iota + 1
	file
)

var outputFromText = map[string]output{
	"STDOUT": stdout,
	"FILE":   file,
}

type Writer struct {
	Rotator

	mx     sync.Mutex
	file   *os.File
	output output
}

func (w *Writer) Write(b []byte) (n int, err error) {
	if w.file == nil {
		return 0, errors.New("file is not set")
	}

	if w.output == 0 {
		return 0, errors.New("output is not set")
	}

	err = w.rotate()
	if err != nil {
		return
	}

	if char := b[len(b)-1]; char != LF && char != CR {
		b = append(b, '\n')
	}

	w.mx.Lock()
	defer w.mx.Unlock()
	return w.file.Write(b)
}

func (w *Writer) rotate() error {
	if w.output == stdout {
		return nil
	}

	oversized, err := w.Oversized(w.file)
	if err != nil {
		return err
	}

	if !oversized {
		return nil
	}

	f, err := w.Rotate(w.file)
	if err != nil {
		return nil
	}

	w.file = f
	return nil
}
