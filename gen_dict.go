package main

import (
	"encoding/json"
	"goDict/configs"
	"goDict/models"
	"goDict/services"
	"log/slog"
)

// getDatabaseInfo 获取数据库信息
func getDatabaseInfo(dbConfig *configs.DatabaseConfig) (*models.DatabaseInfo, error) {
	// 初始化数据库连接
	db, err := configs.InitDatabase(dbConfig)
	if err != nil {
		slog.Error("数据库连接失败", "error", err)
		return nil, err
	}

	// 生成数据库字典服务
	dbDictService := services.NewDbDictService(db)

	return dbDictService.GetDatabaseInfo(dbConfig)
}

// generateDict 生成字典
func generateDict(dbConfig *configs.DatabaseConfig, saveDirPath string, format string) (string, error) {
	// 初始化数据库连接
	db, err := configs.InitDatabase(dbConfig)
	if err != nil {
		slog.Error("数据库连接失败", "error", err)
		return "", err
	}

	// 是否启用数据库模型生产
	if DB_GEN {
		// 使用GeneratorService生成模型
		generatorService := services.NewGeneratorService(db)
		// 生成实体类
		generatorService.GenerateModels("./dao/generate", "./models", "./services")
	}

	// 生成数据库字典服务
	dbDictService := services.NewDbDictService(db)

	// 生成数据库字典
	pathList, err := dbDictService.BuildAll(dbConfig, saveDirPath, format, true)
	if err != nil {
		slog.Error("生成失败", "error", err)
	} else {
		slog.Info("生成成功", "pathList", pathList)
	}

	bytes, err := json.Marshal(pathList)

	return string(bytes), err
}
