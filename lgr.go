package lgr

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type level int

const (
	errorlvl level = iota + 1
	warnlvl
	infolvl
	debuglvl
	tracelvl
)

const timestampLayout = "2006/01/02 15:04:05"

var (
	lvlText = map[level]string{
		errorlvl: "ERROR",
		warnlvl:  "WARN",
		infolvl:  "INFO",
		debuglvl: "DEBUG",
		tracelvl: "TRACE",
	}

	lvlFromText = map[string]level{
		"ERROR": errorlvl,
		"WARN":  warnlvl,
		"INFO":  infolvl,
		"DEBUG": debuglvl,
		"TRACE": tracelvl,
	}
)

func (l level) text() string { return lvlText[l] }

type (
	Logger struct {
		writer *Writer
		level  level
		prefix []string
	}

	// todo: 0 max size, 0 backups, non-existing path etc
	// todo: prettify
	// todo: prefix separator
	Config struct {
		Level           string `json:"level"`            // one of: ["ERROR", "WARN", "INFO", "DEBUG", "TRACE"]
		Output          string `json:"output"`           // one of: ["STDOUT", "FILE"]
		TimestampLayout string `json:"timestamp_layout"` // record prefix time format, default: "2006/01/02 15:04:05"

		// below options are ignored for STDOUT output type
		Path       string `json:"path"`            // relative path to logs directory
		FNameFmt   string `json:"filename_format"` // file name time format with extension (e.g. "2006-Jan-02T15:04:05.999999999.log")
		MaxSizeKB  int    `json:"max_size_kb"`     // file max size
		MaxBackups int    `json:"max_backups"`     // how many files to keep
	}
)

// NewLogger constructor
func NewLogger(config *Config) (l *Logger, err error) {
	lvl, _ := lvlFromText[config.Level]
	if lvl == 0 {
		lvl = tracelvl // trace by default
	}

	w := new(Writer)
	w.output, _ = outputFromText[config.Output]

	switch w.output {
	case stdout:
		w.file = os.Stdout
	case file:
		w.Rotator, err = NewRotator(config)
		if err != nil {
			return nil, err
		}

		w.file, err = w.Rotator.LatestOrNew()
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("illegal output type")
	}

	return &Logger{writer: w, level: lvl, prefix: []string{}}, nil
}

func (l *Logger) Fork(prefix string) *Logger {
	return &Logger{
		writer: l.writer,
		level:  l.level,
		prefix: append(l.prefix, prefix),
	}
}

// TODO: REFACTOR
func (l *Logger) print(level level, v ...interface{}) {
	now := time.Now()

	if !l.shouldPrint(level) {
		return
	}

	var str string

	if len(l.prefix) == 0 {
		str = fmt.Sprintf("%s %s %s", now.Format(timestampLayout), prettify(level.text()), fmt.Sprintln(v...))
	} else {
		str = fmt.Sprintf("%s %s %s: %s", now.Format(timestampLayout), prettify(level.text()), strings.Join(l.prefix, " | "), fmt.Sprintln(v...))
	}

	_, err := l.writer.Write([]byte(str))
	if err != nil {
		fmt.Println(err)
	}
}

// TODO: REFACTOR
func (l *Logger) printf(level level, format string, v ...interface{}) {
	now := time.Now()

	if !l.shouldPrint(level) {
		return
	}

	var str string

	if len(l.prefix) == 0 {
		str = fmt.Sprintf("%s %s %s", now.Format(timestampLayout), prettify(level.text()), fmt.Sprintf(format, v...))
	} else {
		str = fmt.Sprintf("%s %s %s: %s", now.Format(timestampLayout), prettify(level.text()), strings.Join(l.prefix, " | "), fmt.Sprintf(format, v...))
	}

	_, err := l.writer.Write([]byte(str))
	if err != nil {
		fmt.Println(err)
	}
}

func (l *Logger) shouldPrint(level level) bool {
	return l.level >= level
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
