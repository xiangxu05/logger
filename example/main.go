package main

import (
	"time"

	"logger_test/test"

	"github.com/xiangxu05/logger"
)

func main() {
	// 配置日志选项
	cfg := logger.Config{
		MinLevel:      logger.DEBUG,
		Format:        logger.FormatPlain, // 可选 logger.FormatJSON
		Targets:       logger.OutputConsole | logger.OutputFile,
		LogPath:       "logs/app.log",
		AllowedPrefix: []string{"main"}, // 仅白名单包名前缀日志额外输出到 logs_allowed/allowed.log
	}

	// 获取全局 Logger 单例
	log := logger.GetLoggerInstance(cfg)

	// 捕获程序panic并记录
	defer logger.RecoverAndLogPanic()
	// 程序退出时关闭Logger，确保日志写入完成
	defer func() {
		time.Sleep(time.Millisecond * 100)
		log.Close()
	}()

	// 记录各种等级日志
	log.Debug("调试信息")
	log.Info("服务启动成功")
	log.Warn("警告信息")
	log.Error("错误信息")

	// 模拟 panic，测试自动捕获
	// panic("测试 panic 捕获")

	// 程序内部函数使用
	myfunc()
	test.TestFunc()
}

func myfunc() {
	log := logger.GetLoggerInstance()
	log.Info("内部函数使用")
}
