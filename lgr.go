package lgr

import (
	"errors"
	"fmt"
	"log"
	"os"
)

type (
	Logger struct {
		Rotator Rotator

		level  level
		output output
		file   *os.File
		logger *log.Logger

		// prefix []string
		// forks  []*Logger // contains children and parent forks
	}

	// todo: 0 max size, 0 files, non-existing path etc
	Config struct {
		Level  string `json:"level"`  // one of: ["ERROR", "WARN", "INFO", "DEBUG", "TRACE"]
		Output string `json:"output"` // one of ["STDOUT", "FILE"]
		// todo: prettify

		// below options are ignored for STDOUT output type
		Path       string `json:"path"`            // path to logs directory
		FNameFmt   string `json:"filename_format"` // file name time format (e.g. "2006-Jan-02T15:04:05.999999999.log")
		MaxSizeKB  int    `json:"max_size_kb"`     // file max size
		MaxBackups int    `json:"max_backups"`     // how many files to keep
	}
)

// NewLogger constructor
func NewLogger(config *Config) (l *Logger, err error) {
	var nl = func(f *os.File) *log.Logger {
		return log.New(f, "", log.LstdFlags)
	}

	l = &Logger{
		level:  levelFromText[config.Level],
		output: outputFromText[config.Output],
	}

	// var err error
	switch l.output {
	case stdout:
		l.file = os.Stdout
		l.logger = nl(os.Stdout)
	case file:
		l.Rotator, err = NewRotator(config)
		if err != nil {
			return nil, err
		}

		l.file, err = l.Rotator.Latest()
		if err != nil {
			return nil, err
		}

		l.logger = nl(l.file)
	default:
		return nil, errors.New("illegal output type")
	}

	return l, nil
}

// func (l *Logger) Fork(prefix string) *Logger {
// 	l2 := &Logger{
// 		Rotator: l.Rotator,
// 		level:   l.level,
// 		output:  l.output,
// 		file:    l.file,
// 		logger:  l.logger,
// 		prefix:  l.prefix,
// 		forks:   append(l.forks, l),
// 	}
//
// 	l2.prefix = append(l.prefix, prefix)
// 	return l2
// }

func (l *Logger) print(level level, v ...interface{}) {
	if !l.shouldPrint(level) {
		return
	}

	l.preprint()
	l.logger.Printf(fmt.Sprintf("%s %s", prettify(level.text()), fmt.Sprintln(v...)))
}

func (l *Logger) printf(level level, format string, v ...interface{}) {
	if !l.shouldPrint(level) {
		return
	}

	l.preprint()
	l.logger.Println(fmt.Sprintf("%s %s", prettify(level.text()), fmt.Sprintf(format, v...)))
}

func (l *Logger) shouldPrint(level level) bool {
	return l.level >= level
}

func (l *Logger) preprint() {
	if l.output == stdout {
		return
	}

	oversized, err := l.Rotator.Oversized(l.file)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !oversized {
		return
	}

	// for _, l := range l.forks {
	// 	f, err := l.Rotator.Rotate(l.file)
	//
	// 	if err != nil {
	// 		continue
	// 	}
	//
	// 	l.logger.SetOutput(f)
	// 	l.file = f
	// }
	//
	// f, err := l.Rotator.Rotate(l.file)
	// if err != nil {
	// 	return
	// }

	f, err := l.Rotator.Rotate(l.file)
	if err != nil {
		fmt.Println(err)
		return
	}

	l.file = f
	l.logger.SetOutput(f)
}

func (l *Logger) Error(i ...interface{})            { l.print(errorlvl, i...) }
func (l *Logger) Errorf(s string, i ...interface{}) { l.printf(errorlvl, s, i...) }
func (l *Logger) Warn(i ...interface{})             { l.print(warnlvl, i...) }
func (l *Logger) Warnf(s string, i ...interface{})  { l.printf(warnlvl, s, i...) }
func (l *Logger) Info(i ...interface{})             { l.print(infolvl, i...) }
func (l *Logger) Infof(s string, i ...interface{})  { l.printf(infolvl, s, i...) }
func (l *Logger) Debug(i ...interface{})            { l.print(debuglvl, i...) }
func (l *Logger) Debugf(s string, i ...interface{}) { l.printf(debuglvl, s, i...) }

const chars = 5

func prettify(level string) string {
	if len(level) < chars {
		return level + " "
	}

	return level
}
