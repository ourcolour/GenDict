package main

import (
	"errors"
	"goDict/configs"
	"goDict/services"
	"log/slog"
)

func generateDb(dbConfig *configs.DatabaseConfig, saveDirPath string, format string) (string, error) {
	// 初始化数据库配置
	if nil == dbConfig {
		return "", errors.New("数据库连接错误")
	}

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

	/*
		// 上下文
		ctx := context.Background()
		// 查询
		albumService := services.NewAlbumService(db)
		query := &models.Album{
			Name: StringPtr("YST-277"),
		}
		queryResult, err := albumService.SelectByQuery(ctx, query, models.NewQueryOption())
		if err != nil {
			slog.Error("查询失败", "error", err)
		}

		for _, item := range queryResult.Data {
			slog.Info("查询结果", "title", *item.Path)
		}
	*/
	//albumService := services.NewAlbumWebService(db)
	//query := &models.AlbumWeb{}
	//queryResult, err := albumService.SelectByQuery(ctx, query, models.NewQueryOption())
	//if err != nil {
	//	slog.Error("查询失败", "error", err)
	//}
	//
	//for _, album := range queryResult.Data {
	//	slog.Info("查询结果",
	//		slog.Group("album",
	//			"title", *album.Title,
	//			"id", *album.AlbumID, // 假设有 Id 字段
	//		),
	//	)
	//}

	return "", nil
}
