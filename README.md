# lgr

```go
package main

import (
	"fmt"
	"log"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/zzell/lgr"
)

// This example will create "logs" directory in project root and 
// will add new file in format "2006-01-02T15:04:05Z07:00.log".
// At the end it will print file's content to stdout
func main() {
	l, err := lgr.NewLogger(&lgr.Config{
		Level:        "DEBUG",
		Output:       "FILE",
		TimestampFmt: time.RFC3339Nano,
		Separator:    ": ",
		Path:         "logs",
		FNameFmt:     time.RFC3339 + ".log",
		MaxSizeKB:    20_000, // 20MB
		MaxBackups:   5,
	})

	if err != nil {
		log.Fatal(err)
	}

	l.Error("this is error log")

	l2 := l.Fork(stack)

	l2.Warnf("this is warning %s", "with prefix")

	l3 := l2.Fork(func() string { return fmt.Sprintf("NUM_GOROUTINE[%d]", runtime.NumGoroutine()) })

	l3.Info("this is info string with two prefixes")

	ss, _ := l3.Tail(100)

	fmt.Print(strings.Join(ss, "\n"))

	// 2020-09-05T12:53:37.412923+03:00 ERROR this is error log
	// 2020-09-05T12:53:37.41321+03:00 WARN main/main.go:35: this is warning with prefix
	// 2020-09-05T12:53:37.413246+03:00 INFO main/main.go:39: NUM_GOROUTINE[1]: this is info string with two prefixes
}

func stack() string {
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		return "no_stack"
	}

	return fmt.Sprintf("%s:%d", strings.TrimPrefix(file, filepath.Dir(path.Join(path.Dir(file)))+"/"), line)
}
```