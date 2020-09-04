package lgr

import (
	"errors"
	"os"
	"sync"
)

type output int

const (
	stdout output = iota + 1
	file
)

var stroutput = map[string]output{
	"STDOUT": stdout,
	"FILE":   file,
}

type Writer struct {
	rotator Rotator
	mx      sync.Mutex
	file    *os.File
	output  output
}

func NewWriter(rotator Rotator, out output) (w *Writer, err error) {
	var f *os.File

	if rotator == nil && out == file {
		return nil, errors.New("invalid writer args: FILE output requires Rotator's instance")
	}

	if rotator == nil {
		f = os.Stdout
	} else {
		f, err = rotator.File()
		if err != nil {
			return nil, err
		}
	}

	return &Writer{
		rotator: rotator,
		mx:      sync.Mutex{},
		file:    f,
		output:  out,
	}, err
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

	oversized, err := w.rotator.Oversized(w.file)
	if err != nil {
		return err
	}

	if !oversized {
		return nil
	}

	f, err := w.rotator.Rotate(w.file)
	if err != nil {
		return nil
	}

	w.file = f
	return nil
}
