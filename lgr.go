package lgr

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// TODO: Archive old log files

type Level int

const (
	tracelvl Level = iota + 1
	debuglvl
	infolvl
	warnlvl
	errorlvl
)

const (
	defaultLevel     = tracelvl
	defaultOutput    = stdout
	defaultTimestamp = "2006/01/02 15:04:05"
	defaultSeparator = ": "
)

var (
	lvlstr = map[Level]string{
		tracelvl: "TRACE",
		debuglvl: "DEBUG",
		infolvl:  "INFO",
		warnlvl:  "WARN",
		errorlvl: "ERROR",
	}

	strlvl = map[string]Level{
		"TRACE": tracelvl,
		"DEBUG": debuglvl,
		"INFO":  infolvl,
		"WARN":  warnlvl,
		"ERROR": errorlvl,
	}
)

func (l Level) String() string { return lvlstr[l] }

type (
	PrefixFn func() string

	Logger struct {
		Writer    *Writer
		Level     Level
		Prefix    []PrefixFn
		StampFmt  string
		Separator string
	}

	Config struct {
		Level        string `json:"level"`            // one of: ["ERROR", "WARN", "INFO", "DEBUG", "TRACE"] (default: "TRACE")
		Output       string `json:"output"`           // one of: ["STDOUT", "FILE"] (default: "STDOUT")
		TimestampFmt string `json:"timestamp_format"` // timestamp format (default: "2006/01/02 15:04:05")
		Separator    string `json:"prefix_separator"` // prefix separator (default: ": ")

		// below options are not needed for STDOUT logging
		Path       string `json:"path"`            // relative path to logs directory (default: ".")
		FNameFmt   string `json:"filename_format"` // file name time format with extension (default: "2006-Jan-02T15:04:05.999999999.log")
		MaxSizeKB  int    `json:"max_size_kb"`     // file max size before rotation (default: 1000)
		MaxBackups int    `json:"max_backups"`     // how many files to keep (default: 1)
	}
)

func NewLogger(config *Config) (*Logger, error) {
	var (
		lvl     Level
		ok      bool
		err     error
		out     output
		writer  *Writer
		rotator Rotator = nil
	)

	if config == nil {
		config = &Config{}
	}

	if config.Level == "" {
		lvl = defaultLevel
	} else {
		lvl, ok = strlvl[config.Level]
		if !ok {
			return nil, fmt.Errorf("invalid log level %q", config.Level)
		}
	}

	if config.Output == "" {
		out = defaultOutput
	} else {
		out, ok = stroutput[config.Output]
		if !ok {
			return nil, fmt.Errorf("invalid output type %q", config.Output)
		}
	}

	if out == file {
		rotator, err = NewRotator(config.Path, config.FNameFmt, config.MaxSizeKB, config.MaxBackups)
		if err != nil {
			return nil, err
		}
	}

	writer, err = NewWriter(rotator, out)
	if err != nil {
		return nil, err
	}

	timestamp := config.TimestampFmt
	if timestamp == "" {
		timestamp = defaultTimestamp
	}

	separator := config.Separator
	if separator == "" {
		separator = defaultSeparator
	}

	return &Logger{
		Writer:    writer,
		Level:     lvl,
		Prefix:    []PrefixFn{},
		StampFmt:  timestamp,
		Separator: separator,
	}, nil
}

func (l *Logger) Fork(fn PrefixFn) *Logger {
	return &Logger{
		Writer:    l.Writer,
		Level:     l.Level,
		Prefix:    append(l.Prefix, fn),
		StampFmt:  l.StampFmt,
		Separator: l.Separator,
	}
}

const noPrefixLogFmt = "%s %s %s"   // timestamp level log
const prefixLogFmt = "%s %s %s: %s" // timestamp level prefix: log

func (l *Logger) fmt(level Level, format string, v ...interface{}) []byte {
	var (
		now    = time.Now()
		logfmt string
		values = []interface{}{
			now.Format(l.StampFmt),
			level.String(),
		}
	)

	if len(l.Prefix) == 0 {
		logfmt = noPrefixLogFmt
	} else {
		logfmt = prefixLogFmt

		prfxs := make([]string, 0, len(l.Prefix))
		for _, p := range l.Prefix {
			prfxs = append(prfxs, p())
		}

		values = append(values, strings.Join(prfxs, l.Separator))
	}

	if format != "" {
		values = append(values, fmt.Sprintf(format, v...))
	} else {
		values = append(values, fmt.Sprintln(v...))
	}

	return []byte(fmt.Sprintf(logfmt, values...))
}

func (l *Logger) print(level Level, v ...interface{}) {
	if !l.shouldPrint(level) {
		return
	}

	_, err := l.Writer.Write(l.fmt(level, "", v...))
	if err != nil {
		fmt.Println(err)
	}
}

func (l *Logger) printf(level Level, format string, v ...interface{}) {
	if !l.shouldPrint(level) {
		return
	}

	_, err := l.Writer.Write(l.fmt(level, format, v...))
	if err != nil {
		fmt.Println(err)
	}
}

func (l *Logger) Tail(lines int) ([]string, error) {
	if l.Writer.output == stdout {
		return nil, errors.New("STDOUT does not support tailing")
	}

	return TailMany(l.Writer.rotator.Files(), lines)
}

func (l *Logger) Error(i ...interface{})            { l.print(errorlvl, i...) }
func (l *Logger) Errorf(s string, i ...interface{}) { l.printf(errorlvl, s, i...) }
func (l *Logger) Warn(i ...interface{})             { l.print(warnlvl, i...) }
func (l *Logger) Warnf(s string, i ...interface{})  { l.printf(warnlvl, s, i...) }
func (l *Logger) Info(i ...interface{})             { l.print(infolvl, i...) }
func (l *Logger) Infof(s string, i ...interface{})  { l.printf(infolvl, s, i...) }
func (l *Logger) Debug(i ...interface{})            { l.print(debuglvl, i...) }
func (l *Logger) Debugf(s string, i ...interface{}) { l.printf(debuglvl, s, i...) }

func (l *Logger) shouldPrint(level Level) bool {
	return l.Level <= level
}
