# log
日志模块，简单好用

# 部分使用示例

func main(){

	//函数退出时等待日志模块结束(异步模式,需要等待日模块正常结束)
	defer log.Wait()

	//捕获异常进行逻辑处理:捕捉接下来执行是的语句异常
	defer log.Exception(func(stack string, e error) {
		log.Error(stack)
	})

	//捕获异常进行逻辑处理:捕捉try中函数语句块中的异常
	log.Try(func() {
		var a[]int
		a[0] = 1
	}, func(stack string, e error) {
		log.InfoF(stack)
	})

	//设置日志等级  默认 LOG_LEVEL_DEBUG
	log.SetLevel(log.LOG_LEVEL_DEBUG)

	//设置是否异步  默认为true
	log.SetAsync(true)
	//默认日志输出添加文件输出
	log.AddHandler(log.NewLogFileWriter("stdout.log.%Y-%m-%d"))

	//设置自定义格式化头部信息 不设置,就以默认格式输出
	log.SetFormatHeader(func(buf *log.Buffer, level string, line int, file string, dt log.DateTime) {
		buf.AppendBytes('[')
		buf.AppendString(level)
		buf.AppendBytes(' ')
		buf.AppendString(dt.YmdHMS())
		buf.AppendBytes(']')
		dt = nil
	})

	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")
	log.Fatal("fatal")

	log.DebugF("debug %v", 1)
	log.InfoF("info %v", 2)
	log.WarnF("warn %v", 2)
	log.ErrorF("error %v", 4)
	log.FatalF("fatal %v", 5)

	//新建file logger
	w2 := log.NewLogFileWriter("output.log.%Y-%m-%d")
	logger := log.NewLogger(log.LOG_LEVEL_DEBUG, false, w2)
	logger.InfoF("test %v", 1)
	}
