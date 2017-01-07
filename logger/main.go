package logger

// logger is how log messages are processed in nugget.

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// Logger encapsulates a destination and setting for log messages.
type Logger struct {
	logSync     sync.Mutex
	mainOutput  io.Writer
	errorOutput io.Writer
}

//Info writes a log line to the console + backlog of type info.
func (l *Logger) Info(module string, content ...interface{}) {
	l.writeLogLine(false, formatLogPrefix(module, "I", cyan(), content...))
	publishLogMessage("I", module, content...)
}

//Warning writes a log line to the console + backlog of type warning.
func (l *Logger) Warning(module string, content ...interface{}) {
	l.writeLogLine(true, formatLogPrefix(module, "W", yellow(), content...))
	publishLogMessage("W", module, content...)
}

//Error writes a log line to the console + backlog of type error.
func (l *Logger) Error(module string, content ...interface{}) {
	l.writeLogLine(true, formatLogPrefix(module, "E", red(), content...))
	publishLogMessage("E", module, content...)
}

//Fatal writes a log line to the console + backlog of type fatal, then terminates the program.
func (l *Logger) Fatal(module string, content ...interface{}) {
	l.writeLogLine(true, formatLogPrefix(module, "F", red(), content...))
	publishLogMessage("F", module, content...)
	os.Exit(1)
}

func formatLogPrefix(module, messagePrefix, prefixColor string, content ...interface{}) string {
	c := fmt.Sprint(content...)
	module = strings.ToUpper(module)
	if module != "" {
		return prefixColor + "[" + messagePrefix + "] " + blue() + "[" + module + "] " + clear() + c
	}
	return prefixColor + "[" + messagePrefix + "] " + clear() + c
}

func (l *Logger) writeLogLine(isError bool, inp string) {
	l.logSync.Lock()
	defer l.logSync.Unlock()
	if isError {
		l.mainOutput.Write([]byte(inp + "\n"))
	} else {
		l.mainOutput.Write([]byte(inp + "\n"))
	}
}

// New returns a new logger, ready to use.
func New(main, err io.Writer) *Logger {
	return &Logger{
		mainOutput:  main,
		errorOutput: err,
	}
}

func publishLogMessage(msgType, module string, content ...interface{}) {
	//We can implement this if we want other systems to recieve logs

	//c := fmt.Sprint(content...)
	//publishMessage(module, msgType, c)
	//addToBacklog(module, msgType, c)
}
