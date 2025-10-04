package main

import (
	"fyne.io/fyne/v2/app"
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

	// 初级化i18n
	if err := utils.InitI18n(); nil != err {
		slog.Error("Failed to initializing i18n.", "error", err)
		panic(err)
	}
	slog.Info("App is now running ...")

	// 初始化GUI
	appInstance := app.New()
	mainView := NewMainView(&appInstance)
	mainView.Show()

	slog.Info("App exited ...")
}
