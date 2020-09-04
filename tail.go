package lgr

import (
	"io"
	"os"
)

// http://www.asciitable.com/
const (
	LF byte = 10 // Line Feed '\n'
	CR byte = 13 // Carriage Return '\r'
)

// TailMany the same as Tail but reads from multiple files.
// Files should be sorted in ASC.
func TailMany(files []string, n int) (ss []string, err error) {
	if len(files) == 0 || n == 0 {
		return ss, nil
	}

	for i := 0; ; i++ {
		if len(files) == i {
			return ss, nil
		}

		f, err := os.Open(files[i])
		if err != nil {
			return nil, err
		}

		lines, err := Tail(f, n)
		if err != nil {
			return nil, err
		}

		_ = f.Close()

		ss = append(lines, ss...)
		if len(lines) < n {
			n -= len(lines)
			continue
		}

		return ss, nil
	}
}

// Tail reads last N lites from file
func Tail(file *os.File, n int) (records []string, err error) {
	stat, err := file.Stat()
	if err != nil {
		return
	}

	if stat.Size() == 0 {
		return nil, nil
	}

	line := make([]byte, 0)

	for cursor, index := -1, 0; ; cursor-- {
		_, err = file.Seek(int64(cursor), io.SeekEnd)
		if err != nil {
			return
		}

		var char = make([]byte, 1)
		_, err = file.Read(char)
		if err != nil {
			return
		}

		// newline
		if char[0] == LF || char[0] == CR {
			index++

			// ignore last LF/CR
			if index == 1 && cursor == -1 {
				continue
			}

			records = append([]string{string(line)}, records...)

			// we are done
			if index == n {
				return
			}

			line = make([]byte, 0)
			continue
		}

		line = append(char, line...)

		// beginning of the file, but not enough lines
		if int64(cursor) == -stat.Size() {
			return append([]string{string(line)}, records...), nil
		}
	}
}
