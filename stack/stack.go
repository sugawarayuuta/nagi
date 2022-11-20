package stack

import "runtime"

type Frame struct {
	Fl    string
	Ln    int
	Count uintptr
}

type Stacktrace struct {
	Err    error
	Stack  []uintptr
	Frames []Frame
}

const (
	MAX = 10
)

func Stack(err error, skip int) *Stacktrace {
	if err == nil {
		return nil
	}
	stack := make([]uintptr, MAX)
	len := runtime.Callers(2+skip, stack[:])
	stacktrace := new(Stacktrace)
	stacktrace.Err = err
	stacktrace.Stack = stack[:len]
	stacktrace.Frames = stacktrace.frame()
	return stacktrace
}

func (err Stacktrace) frame() []Frame {
	err.Frames = make([]Frame, len(err.Stack))
	for index, ptr := range err.Stack {
		frame := Frame{Count: ptr}
		err.Frames[index].Fl, err.Frames[index].Ln = runtime.FuncForPC(frame.Count).FileLine(ptr - 1)
	}
	return err.Frames
}
