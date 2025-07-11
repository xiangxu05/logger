package test

import "github.com/xiangxu05/logger"

func TestFunc() {
	log := logger.GetLoggerInstance()
	log.Info("内部函数使用")
}
