package lgr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	defaultPath       = "."
	defaultFormat     = "2006-Jan-02T15:04:05.999999999.log"
	defaultMaxSizeKB  = 1000
	defaultMaxBackups = 1
)

type (
	// Rotator describes log files rotation, backup, cleaning and tailing
	Rotator interface {
		Rotate(*os.File) (*os.File, error)
		Oversized(*os.File) (bool, error)
		New() (*os.File, error)
		File() (*os.File, error) // todo: bad name
		Clean() error
		Files() []string
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
//
// Reads directory from config, parses it's files as a time format and sorts.
// If directory does not exist it will be created as well as a file.
func NewRotator(path, format string, maxSizeKB, maxBackups int) (Rotator, error) {
	r := &rotator{
		fs:         make([]time.Time, 0),
		path:       path,
		format:     format,
		maxSizeKB:  maxSizeKB,
		maxBackups: maxBackups,
	}

	if r.path == "" {
		r.path = defaultPath
	}

	if r.format == "" {
		r.format = defaultFormat
	}

	if r.maxSizeKB == 0 {
		r.maxSizeKB = defaultMaxSizeKB
	}

	if r.maxBackups == 0 {
		r.maxBackups = defaultMaxBackups
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

func (r *rotator) Files() []string {
	var files = make([]string, 0, r.Len())

	for _, f := range r.fs {
		files = append(files, filepath.Join(r.path, f.Format(r.format)))
	}

	return files
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

func (r *rotator) File() (*os.File, error) {
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
