package errors

import "strings"

type StackSource struct {
	SourceFileName     string `json:"source"`
	SourceFunctionName string `json:"func"`
}

type state struct {
	b []byte
}

// Write implement fmt.Formatter interface.
func (s *state) Write(b []byte) (n int, err error) {
	s.b = b
	return len(b), nil
}

// Width implement fmt.Formatter interface.
func (s *state) Width() (wid int, ok bool) {
	return 0, false
}

// Precision implement fmt.Formatter interface.
func (s *state) Precision() (prec int, ok bool) {
	return 0, false
}

// Flag implement fmt.Formatter interface.
func (s *state) Flag(c int) bool {
	return false
}

func frameField(f StackFrame, s *state, c rune) string {
	f.Format(s, c)
	return string(s.b)
}

func MarshalStack(err error) interface{} {
	type stackTracer interface {
		StackFrames() []StackFrame
	}

	var sterr stackTracer
	var ok bool
	for err != nil {
		sterr, ok = err.(stackTracer)
		if ok {
			break
		}

		u, ok := err.(interface {
			Unwrap() error
		})
		if !ok {
			return nil
		}

		err = u.Unwrap()
	}
	if sterr == nil {
		return nil
	}

	st := sterr.StackFrames()
	s := &state{}
	errLength := len(strings.Split(err.(*Error).Error(), " -- "))

	objOut := make([]StackSource, errLength)

	for i, frame := range st[:errLength] {
		objOut[i] = StackSource{
			SourceFileName:     frameField(frame, s, 's'),
			SourceFunctionName: frameField(frame, s, 'n'),
		}
	}

	return objOut
}
