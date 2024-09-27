package eventbus

import (
	"fmt"
	"runtime"
)

func Error(err error) error {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return fmt.Errorf("%s:%d: %w", frame.Function, frame.Line, err)
}
