package log

import "fmt"

const oSkip = logSkip

func SetLevel(level int) {
	log.SetLevel(level)
}

func AddHandler(handler logWriter) {
	log.AddHandler(handler)
}

func SetAsync(async bool) {
	log.SetAsync(async)
}

func Debug(v ...interface{}) {
	log.output(LOG_LEVEL_DEBUG, fmt.Sprint(v...))
}

func Info(v ...interface{}) {
	log.output(LOG_LEVEL_INFO, fmt.Sprint(v...))
}

func Warn(v ...interface{}) {
	log.output(LOG_LEVEL_WARN, fmt.Sprint(v...))
}

func Error(v ...interface{}) {
	log.output(LOG_LEVEL_ERROR, fmt.Sprint(v...))
}

func Fatal(v ...interface{}) {
	log.output(LOG_LEVEL_FATAL, fmt.Sprint(v...))
}

func DebugF(format string, v ...interface{}) {
	log.output(LOG_LEVEL_DEBUG, fmt.Sprintf(format, v...))
}

func InfoF(format string, v ...interface{}) {
	log.output(LOG_LEVEL_INFO, fmt.Sprintf(format, v...))
}

func WarnF(format string, v ...interface{}) {
	log.output(LOG_LEVEL_WARN, fmt.Sprintf(format, v...))
}

func ErrorF(format string, v ...interface{}) {
	log.output(LOG_LEVEL_ERROR, fmt.Sprintf(format, v...))
}

func FatalF(format string, v ...interface{}) {
	log.output(LOG_LEVEL_FATAL, fmt.Sprintf(format, v...))
}

func SetFormatHeader(formatHeader func(buf *Buffer, level string, line int, file string, dt DateTime)) {
	log.SetFormatHeader(formatHeader)
}

func Wait() { //等待异步日志模块退出
	loggerMgr.Wait()
	log.Wait()
}
