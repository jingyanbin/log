package log

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

const exceptionSkip = 5
//const exceptionSkip = 6

func newError(format string, a ...interface{}) error {
	return errors.New(fmt.Sprintf(format, a...))
}

func toError(r interface{}) (err error) {
	switch x := r.(type) {
	case string:
		err = errors.New(x)
	case error:
		err = x
	default:
		err = newError("unknown error: %v", x)
	}
	return
}

func callerShort(skip int) (file string, line int) {
	var ok bool
	_, file, line, ok = runtime.Caller(skip)
	if ok {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
	} else{
		file = "???"
		line = 0
	}
	return
}

func itoa(dst *[]byte, i int, w int) {
	var b = [20]byte{48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48}
	pos := 19
	for i > 9 {
		m := i / 10
		b[pos] = byte(48 + i - m*10)
		i = m
		pos--
		w--
	}
	b[pos] = byte(48 + i)

	pos2 := 20 - w
	if pos2 > pos {
		*dst = append(*dst, b[pos:]...)
	} else {
		*dst = append(*dst, b[pos2:]...)
	}
}

func formatStack(name, file string, line int, err string, stack[]byte) *Buffer{
	buf := getBuffer(160 + len(stack) + len(name))
	buf.AppendStrings("exception panic: ", err, " from ", file, ":")
	buf.AppendInt(line, 0)
	buf.AppendStrings("(", name, ")\n")
	buf.AppendBytes(stack...)
	return buf
}

var reLine = regexp.MustCompile(`^panic\([a-z 0-9]+,\s*[a-z 0-9]+\)$`)

func callerLineStack(stack string) (name string, file string, line int, success bool) {
	stackLines := strings.Split(stack, "\n")
	max := len(stackLines)
	for i, v := range stackLines{
		if reLine.MatchString(v){
			if i + 3 < max{
				fls := strings.Trim(stackLines[i + 3], "\t")
				fileLines := strings.Split(fls, " ")[0]
				index := strings.LastIndex(fileLines, ":")
				if index == -1{
					return
				}
				file = fileLines[:index]
				var err error
				line, err = strconv.Atoi(fileLines[index+1:])
				if err != nil{
					return
				}
				name = stackLines[i + 2]// strings.Split(stackLines[i + 2], "(")[0]
				success = true
				return
			} else {
				return
			}
		}
	}
	return
}

func callerInFunc(skip int) (name string, file string, line int) {
	var pc uintptr
	var ok bool
	pc , file, line, ok = runtime.Caller(skip)
	if ok{
		inFunc := runtime.FuncForPC(pc)
		name = inFunc.Name()
	} else {
		file = "???"
		name = "???"
	}
	return
}

func Exception(catch func(stack string, e error)){
	if err := recover(); err != nil {
		info := debug.Stack()
		name, file, line, success := callerLineStack(string(info))
		if !success{
			name, file, line = callerInFunc(exceptionSkip)
		}
		myErr := toError(err)
		myBuf := formatStack(name, file,line, myErr.Error(), info)
		defer myBuf.free()
		if catch == nil{
			log.output(LOG_LEVEL_FATAL, string(*myBuf))
		} else{
			catch(string(*myBuf), myErr)
		}
	}
}

func Try(f func(), catch func(stack string, e error)){
	defer Exception(catch)
	f()
}