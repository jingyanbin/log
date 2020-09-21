package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	_ "unsafe"
)

const (
	LOG_LEVEL_DEBUG = iota
	LOG_LEVEL_INFO
	LOG_LEVEL_WARN
	LOG_LEVEL_ERROR
	LOG_LEVEL_FATAL
	LOG_LEVEL_OFF
)

const logSkip = 3

var logLevels = []string{"DEBU", "INFO", "WARN", "ERRO", "CRIT"}

type logWriter interface {
	Write(unix int64, level int, file string, line int, content string)
	SetFormatHeader(formatHeader func(buf *Buffer, level string, line int, file string, dt DateTime))
	Close()
}

type logBaseWriter struct {
	time         datetime
	customHeader func(buf *Buffer, level string, line int, file string, dt DateTime)
}

func (my *logBaseWriter) Close() {}

func (my *logBaseWriter) SetFormatHeader(formatHeader func(buf *Buffer, level string, line int, file string, dt DateTime)) {
	my.customHeader = formatHeader
}

func (my *logBaseWriter) formatHeader(buf *Buffer, level string, line int, file string) {
	buf.AppendBytes('[')
	buf.AppendString(level)
	buf.AppendBytes(' ')
	buf.AppendInt(my.time.year, 4)
	buf.AppendBytes('/')
	buf.AppendInt(my.time.month, 2)
	buf.AppendBytes('/')
	buf.AppendInt(my.time.day, 2)
	buf.AppendBytes(' ')
	buf.AppendInt(my.time.hour, 2)
	buf.AppendBytes(':')
	buf.AppendInt(my.time.min, 2)
	buf.AppendBytes(':')
	buf.AppendInt(my.time.sec, 2)
	buf.AppendBytes(' ')
	buf.AppendString(file)
	buf.AppendBytes(':')
	buf.AppendInt(line, 0)
	buf.AppendString("]")
}

func (my *logBaseWriter) write(writer io.Writer, level int, file string, line int, content string) {
	//buf := make([]byte, 0, 40+len(file)+len(content))
	buf := getBuffer(40 + len(file) + len(content))
	defer buf.free()
	if my.customHeader == nil {
		my.formatHeader(buf, logLevels[level], line, file)
	} else {
		my.customHeader(buf, logLevels[level], line, file, &my.time)
	}
	buf.AppendBytes(' ')
	buf.AppendString(content)
	buf.AppendBytes('\n')
	//buf = append(buf, content...)
	//buf = append(buf, '\n')
	n, err := writer.Write(*buf)

	if err != nil {
		fmt.Printf("logBaseWriter write error: %v, n=%v\n", err, n)
	}
}

type logStdWriter struct {
	logBaseWriter
	writer io.Writer
}

func (my *logStdWriter) Write(unix int64, level int, file string, line int, content string) {
	my.time.flushTo(unix)
	my.write(my.writer, level, file, line, content)
}

func NewLogStdWriter(writer io.Writer) *logStdWriter {
	out := &logStdWriter{}
	out.writer = writer
	return out
}

type logFileWriter struct {
	logBaseWriter
	writer            *handleFile
	filePathFormatter string
	mu                sync.Mutex
}

func (my *logFileWriter) NextPathName() (folderPath string, fileName string) {
	return filepath.Split(my.time.format(my.filePathFormatter))
}

func (my *logFileWriter) Write(unix int64, level int, file string, line int, content string) {
	my.mu.Lock()
	defer my.mu.Unlock()
	my.time.flushTo(unix)
	my.writer.SetPathName(my.NextPathName())
	my.write(my.writer, level, file, line, content)
}

func (my *logFileWriter) Close() {
	my.mu.Lock()
	defer my.mu.Unlock()
	my.writer.Close()
}

func NewLogFileWriter(filePathFormatter string) *logFileWriter {
	out := &logFileWriter{}
	out.writer = &handleFile{flag: os.O_WRONLY | os.O_APPEND | os.O_CREATE, perm: 0777}
	if filePathFormatter == "" {
		out.filePathFormatter = filepath.Join(execDir(), "output.log.%Y-%m-%d-%H")
	} else {
		out.filePathFormatter = filePathFormatter
	}
	return out
}

type stdLogger struct {
	level    int
	handlers []logWriter
	queue    *queue
	running  int32
	async    bool
	wg       sync.WaitGroup
}

func (my *stdLogger) output(level int, content string) {
	if level < my.level {
		return
	}
	file, line := callerShort(logSkip)
	my.write(unix(), level, file, line, content)
}

func (my *stdLogger) SetLevel(level int) {
	if level < LOG_LEVEL_DEBUG || level > LOG_LEVEL_OFF {
		return
	}
	my.level = level
}

func (my *stdLogger) AddHandler(handlers ... logWriter) {
	for _, handler := range handlers {
		if handler == nil {
			continue
		}
		my.handlers = append(my.handlers, handler)
	}
}

func (my *stdLogger) SetAsync(async bool) {
	my.async = async
	if async {
		my.start()
	} else {
		my.Wait()
	}
}

func (my *stdLogger) start() {
	if atomic.CompareAndSwapInt32(&my.running, 0, 1) {
		if my.queue == nil {
			my.queue = newQueue()
		}
		my.queue.Open()
		my.wg.Add(1)
		go my.run()
	}
}

func (my *stdLogger) Wait() {
	if atomic.CompareAndSwapInt32(&my.running, 1, 0) {
		my.queue.Close()
		my.wg.Wait()
	}
	my.close()
}

func (my *stdLogger) SetFormatHeader(formatHeader func(buf *Buffer, level string, line int, file string, dt DateTime)) {
	for _, handler := range my.handlers {
		handler.SetFormatHeader(formatHeader)
	}
}

func (my *stdLogger) close() {
	for _, handler := range my.handlers {
		handler.Close()
	}
}

func (my *stdLogger) run() {
	defer Exception(func(stack string, e error) {
		fmt.Println(stack, e)
	})
	defer my.wg.Done()
	var item *logItem
	var ok bool
	var handler logWriter
	var value interface{}
	var closed bool
	for {
		value, closed = my.queue.PopBlock()
		if closed {
			break
		}
		item, ok = value.(*logItem)
		if !ok {
			break
		}
		for _, handler = range my.handlers {
			handler.Write(item.unix, item.level, item.file, item.line, item.content)
		}
		item.free()
	}
}

func (my *stdLogger) write(unix int64, level int, file string, line int, content string) () {
	if my.async {
		item := logItemFree.Get().(*logItem)
		item.unix = unix
		item.level = level
		item.content = content
		item.file = file
		item.line = line
		if my.queue.Push(item) {
			return
		}
	}
	for _, handler := range my.handlers {
		handler.Write(unix, level, file, line, content)
	}
}

func (my *stdLogger) Debug(v ...interface{})                 { my.output(LOG_LEVEL_DEBUG, fmt.Sprint(v...)) }
func (my *stdLogger) Info(v ...interface{})                  { my.output(LOG_LEVEL_INFO, fmt.Sprint(v...)) }
func (my *stdLogger) Warn(v ...interface{})                  { my.output(LOG_LEVEL_WARN, fmt.Sprint(v...)) }
func (my *stdLogger) Error(v ...interface{})                 { my.output(LOG_LEVEL_ERROR, fmt.Sprint(v...)) }
func (my *stdLogger) Fatal(v ...interface{})                 { my.output(LOG_LEVEL_FATAL, fmt.Sprint(v...)) }
func (my *stdLogger) DebugF(format string, v ...interface{}) { my.output(LOG_LEVEL_DEBUG, fmt.Sprintf(format, v...)) }
func (my *stdLogger) InfoF(format string, v ...interface{})  { my.output(LOG_LEVEL_INFO, fmt.Sprintf(format, v...)) }
func (my *stdLogger) WarnF(format string, v ...interface{})  { my.output(LOG_LEVEL_WARN, fmt.Sprintf(format, v...)) }
func (my *stdLogger) ErrorF(format string, v ...interface{}) { my.output(LOG_LEVEL_ERROR, fmt.Sprintf(format, v...)) }
func (my *stdLogger) FatalF(format string, v ...interface{}) { my.output(LOG_LEVEL_FATAL, fmt.Sprintf(format, v...)) }

func NewLogger(level int, async bool, handlers ... logWriter) *stdLogger {
	logger := &stdLogger{}
	logger.SetLevel(level)
	logger.AddHandler(handlers...)
	logger.SetAsync(async)
	runtime.SetFinalizer(logger, (*stdLogger).Wait)
	loggerMgr.Append(logger)
	return logger
}

type loggerManager struct {
	loggers []*stdLogger
	mu      sync.Mutex
}

func (my *loggerManager) Append(logger *stdLogger) {
	my.mu.Lock()
	defer my.mu.Unlock()
	my.loggers = append(my.loggers, logger)
}

func (my *loggerManager) Wait() {
	for _, logger := range my.loggers {
		logger.Wait()
	}
}

type logItem struct {
	unix    int64
	level   int
	content string
	file    string
	line    int
}

func (it *logItem) free() {
	logItemFree.Put(it)
}

var loggerMgr loggerManager

var logItemFree = sync.Pool{New: func() interface{} { return new(logItem) },}

var log = stdLogger{level: LOG_LEVEL_DEBUG, handlers: []logWriter{NewLogStdWriter(os.Stdout)}}

func init() {
	runtime.SetFinalizer(&log, (*stdLogger).Wait)
	log.SetAsync(true)
}
