package lgr

const (
	errorlvl level = iota
	warnlvl
	infolvl
	debuglvl
	tracelvl
)

const (
	stdout output = iota
	file
)

type (
	level  int
	output int
)

var (
	levelText = map[level]string{
		errorlvl: "ERROR",
		warnlvl:  "WARN",
		infolvl:  "INFO",
		debuglvl: "DEBUG",
		tracelvl: "TRACE",
	}

	levelFromText = map[string]level{
		"ERROR": errorlvl,
		"WARN":  warnlvl,
		"INFO":  infolvl,
		"DEBUG": debuglvl,
		"TRACE": tracelvl,
	}

	outputFromText = map[string]output{
		"STDOUT": stdout,
		"FILE":   file,
	}
)

func (l level) text() string {
	return levelText[l]
}
