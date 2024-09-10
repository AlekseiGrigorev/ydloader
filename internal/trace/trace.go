// Copyright 2024 Aleksei Grigorev
// https://aleksvgrig.com, https://github.com/AlekseiGrigorev, aleksvgrig@gmail.com.
// Package define define interfaces, structures and functions for trace applications
package trace

import (
	"fmt"
	"runtime"
)

// Returns current trace stack
func GetTrace() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
}
