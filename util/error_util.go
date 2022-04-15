package util

import (
	"fmt"
	"runtime"
	"strings"
)

var Err *errUtil = &errUtil{}

type errUtil struct {
}

// print stack trace for debug
func (*errUtil) Trace(message string) string {
	var pcs [32]uintptr
	callerList := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var builder strings.Builder
	builder.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:callerList] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		builder.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return builder.String()
}

func (*errUtil) ServerError(code string, message string) error {
	return fmt.Errorf("%s:%s", code, message)
}
