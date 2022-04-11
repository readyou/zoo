package util

import (
	"fmt"
	"runtime"
	"strings"
)

// print stack trace for debug
func Trace(message string) string {
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
