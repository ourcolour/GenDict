package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"
)

var (
	IDX_COMMENT_COUNT = 0
	// 应用名称
	APP_NAME = "GenDict"

	// 开启调试模式
	DEBUG = os.Getenv("DEBUG") == "true"

	// 开启数据库模型生成
	DB_GEN = os.Getenv("DB_GEN") == "true"

	// 基本配置
	WORK_DIR, _ = os.Getwd()

	// 日志等级
	LOG_LEVEL = slog.LevelDebug.String()

	// 日志文件
	LOG_PATH = path.Join(WORK_DIR, "logs", fmt.Sprintf("%s.log", APP_NAME))
)
