package main

import (
	"goDict/utils"
	"log/slog"
)

var (
	StringPtr = utils.StringPtr
)

func main() {
	// 初始化日志
	if err := utils.InitLogger(LOG_LEVEL, "./logs"); nil != err {
		panic(err)
	}

	slog.Info("程序启动...")

	// 初始化GUI
	mainView := NewMainView()
	mainView.Show()

	slog.Info("程序退出...")
}
