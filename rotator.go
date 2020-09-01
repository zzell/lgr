package lgr

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type (
	// Rotator describes log files rotation, backup, cleaning and tailing
	Rotator interface {
		Rotate(*os.File) (*os.File, error)
		Oversized(*os.File) (bool, error)
		New() (*os.File, error)
		LatestOrNew() (*os.File, error)
		Clean() error
		Tail(int) ([][]byte, error)
	}

	rotator struct {
		fs         []time.Time
		path       string
		format     string
		maxSizeKB  int
		maxBackups int
	}
)

// NewRotator constructor
func NewRotator(config *Config) (Rotator, error) {
	r := &rotator{
		fs:         make([]time.Time, 0),
		path:       config.Path,
		format:     config.FNameFmt,
		maxSizeKB:  config.MaxSizeKB,
		maxBackups: config.MaxBackups,
	}

	dir, err := ioutil.ReadDir(r.path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(r.path, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	for _, f := range dir {
		if f.IsDir() {
			continue
		}

		t, err := time.Parse(r.format, f.Name())
		if err != nil {
			continue
		}

		r.fs = append(r.fs, t)
	}

	sort.Sort(r)
	return r, nil
}

func (r *rotator) Len() int {
	return len(r.fs)
}

func (r *rotator) Swap(i, j int) {
	r.fs[i], r.fs[j] = r.fs[j], r.fs[i]
}

func (r *rotator) Less(i, j int) bool {
	return r.fs[i].After(r.fs[j])
}

func (r *rotator) Rotate(f *os.File) (*os.File, error) {
	_ = f.Close()

	// todo: remove clean from here
	if err := r.Clean(); err != nil {
		return nil, err
	}

	return r.New()
}

func (r *rotator) Oversized(f *os.File) (bool, error) {
	stat, err := f.Stat()
	if err != nil {
		return false, err
	}

	return stat.Size()/1024 >= int64(r.maxSizeKB), nil
}

func (r *rotator) New() (*os.File, error) {
	var now = time.Now()

	f, err := os.Create(filepath.Join(r.path, now.Format(r.format)))
	if err != nil {
		return nil, err
	}

	r.fs = append([]time.Time{now}, r.fs...)
	return f, err
}

func (r *rotator) LatestOrNew() (*os.File, error) {
	// directory is empty - creating new file
	if len(r.fs) == 0 {
		return r.New()
	}

	return os.OpenFile(filepath.Join(r.path, r.fs[0].Format(r.format)), os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
}

func (r *rotator) Clean() error {
	i := r.Len() - r.maxBackups
	if i <= 0 {
		return nil
	}

	for ; i > 0; i-- {
		err := os.Remove(filepath.Join(r.path, r.fs[r.Len()-1].Format(r.format)))
		if err != nil {
			return err
		}

		// remove latest file
		r.fs = r.fs[:r.Len()-1]
	}

	return nil
}

func (r *rotator) Tail(n int) ([][]byte, error) {
	if r.Len() == 0 || n == 0 {
		return make([][]byte, 0), nil
	}

	var (
		left = n
		recs = make([][]byte, 0)
	)

	for i := 0; ; i++ {
		if r.Len() == i {
			return recs, nil
		}

		f, err := os.Open(filepath.Join(r.path, r.fs[i].Format(r.format)))
		if err != nil {
			return nil, err
		}

		records, err := tail(f, left)
		if err != nil {
			return nil, err
		}

		_ = f.Close()

		recs = append(records, recs...)
		if len(records) < left {
			left -= len(records)
			continue
		}

		return recs, nil
	}
}

func tail(file *os.File, n int) (records [][]byte, err error) {
	stat, err := file.Stat()
	if err != nil {
		return
	}

	if stat.Size() == 0 {
		return nil, nil
	}

	line := make([]byte, 0)

	// default logger prints newline character at the end of a file
	// https://golang.org/pkg/log/#Logger.Output
	//
	// '-1' - latest char in the file (which is '\n')
	// '-2' - the second char after the last one
	for cursor, index := -2, 0; ; cursor-- {
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
		if char[0] == 10 || char[0] == 13 {
			index++
			records = append([][]byte{line}, records...)

			// enough lines, we are done
			if index == n {
				return
			}

			line = make([]byte, 0)
			continue
		}

		line = append(char, line...)

		// beginning of the file, but not enough lines
		if int64(cursor) == -stat.Size() {
			return append([][]byte{line}, records...), nil
		}
	}
}
