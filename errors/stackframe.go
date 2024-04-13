package errors

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

// A StackFrame contains all necessary information about to generate a line
// in a callstack.
type StackFrame struct {
	// The path to the file containing this ProgramCounter
	File string
	// The LineNumber in that file
	LineNumber int
	// The Name of the function that contains this ProgramCounter
	Name string
	// The Package that contains this function
	Package string
	// The underlying ProgramCounter
	ProgramCounter uintptr
}

var cwd string

func (err *Error) Stack() []byte {
	buf := bytes.Buffer{}

	for _, frame := range err.StackFrames() {
		buf.WriteString(frame.String())
	}

	return buf.Bytes()
}

func (err *Error) StackFrames() []StackFrame {
	if err.frames == nil {
		err.frames = make([]StackFrame, len(err.stack))

		for i, pc := range err.stack {
			err.frames[i] = NewStackFrame(pc)
		}
	}

	return err.frames
}

// NewStackFrame popoulates a stack frame object from the program counter.
func NewStackFrame(pc uintptr) (frame StackFrame) {
	frame = StackFrame{ProgramCounter: pc}
	if frame.Func() == nil {
		return
	}

	frame.Package, frame.Name = packageAndName(frame.Func())

	// pc -1 because the program counters we use are usually return addresses,
	// and we want to show the line that corresponds to the function call
	frame.File, frame.LineNumber = frame.Func().FileLine(pc - 1)

	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	if strings.HasPrefix(frame.File, cwd) {
		frame.File = frame.File[len(cwd)+1:]
	}

	return

}

// Func returns the function that contained this frame.
func (frame *StackFrame) Func() *runtime.Func {
	if frame.ProgramCounter == 0 {
		return nil
	}
	return runtime.FuncForPC(frame.ProgramCounter)
}

// String returns the stackframe formatted in the same way as go does
// in runtime/debug.Stack()
func (frame *StackFrame) String() string {
	str := fmt.Sprintf("%s:%d (0x%x)\n", frame.File, frame.LineNumber, frame.ProgramCounter)

	source, err := frame.sourceLine()
	if err != nil {
		return str
	}

	return str + fmt.Sprintf("\t%s: %s\n", frame.Name, source)
}

// SourceLine gets the line of code (from File and Line) of the original source if possible.
func (frame *StackFrame) SourceLine() (string, error) {
	source, err := frame.sourceLine()
	if err != nil {
		return source, New(err)
	}
	return source, err
}

func (frame *StackFrame) sourceLine() (string, error) {
	if frame.LineNumber <= 0 {
		return "???", nil
	}

	file, err := os.Open(frame.File)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine == frame.LineNumber {
			return string(bytes.Trim(scanner.Bytes(), " \t")), nil
		}
		currentLine++
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "???", nil
}

func packageAndName(fn *runtime.Func) (string, string) {
	name := fn.Name()
	pkg := ""

	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//  runtime/debug.*T·ptrmethod
	// and want
	//  *T.ptrmethod
	// Since the package path might contains dots (e.g. code.google.com/...),
	// we first remove the path prefix if there is one.
	if lastslash := strings.LastIndex(name, "/"); lastslash >= 0 {
		pkg += name[:lastslash] + "/"
		name = name[lastslash+1:]
	}
	if period := strings.Index(name, "."); period >= 0 {
		pkg += name[:period]
		name = name[period+1:]
	}

	name = strings.Replace(name, "·", ".", -1)
	return pkg, name
}

// Format formats the frame according to the fmt.Formatter interface.
//
//	%s    source file
//	%d    source line
//	%n    function name
//	%v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//	%+s   function name and path of source file relative to the compile time
//	      GOPATH separated by \n\t (<funcname>\n\t<path>)
//	%+v   equivalent to %+s:%d
func (f StackFrame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.Func().Name())
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.File)
		default:
			io.WriteString(s, fmt.Sprintf("%s:%d:", f.File, f.LineNumber))
		}
	case 'd':
		line, _ := f.sourceLine()
		io.WriteString(s, line)
	case 'n':
		funcName := strings.Split(f.Func().Name(), "/")
		io.WriteString(s, funcName[len(funcName)-1])
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}
