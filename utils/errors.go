package utils

import (
	"bytes"
	"fmt"
	"runtime"
)

func ErrorStack(skip int, err error) string {
	var buf bytes.Buffer

	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])

	buf.WriteString("Error:\n")
	buf.WriteString(err.Error())

	for _, t := range pcs[0:n] {
		buf.WriteString("\n")
		pc := t - 1
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			buf.WriteString("\nFile: ")
			file, line := fn.FileLine(pc)
			buf.WriteString(file)
			buf.WriteString(fmt.Sprintf(" +%d", line))
			buf.WriteString("\nFunction: ")
			buf.WriteString(fn.Name())
		}
	}

	return buf.String()
}
