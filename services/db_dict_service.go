package services

import (
	"errors"
	"goDict/configs"
	"goDict/models"
	"gorm.io/gorm"
	"log/slog"
	"slices"
	"sort"
)

// 支持的格式
var SUPPORTED_FORMAT = map[string]bool{
	"md":   true,
	"xls":  true,
	"xlsx": true,
}

type DbDictService struct {
	DB *gorm.DB
}

func NewDbDictService(db *gorm.DB) *DbDictService {
	return &DbDictService{
		DB: db,
	}
}

// buildTableInfo 获取表信息
func (this *DbDictService) buildTableInfo(
	databaseName string,
	tableName string,
	tableTypeMap *map[string]string,
	tableColumnInfoMap *map[string][]*models.ColumnInfo,
	indexInfoListMap *map[string][]*models.IndexInfo,
	tableCommemtMap *map[string]string,
) (*models.TableInfo, error) {
	// 获取对象类型
	tableType := (*tableTypeMap)[tableName]
	// 获取索引
	indexInfoList := (*indexInfoListMap)[tableName]
	// 获取表注释
	tableCommemt := (*tableCommemtMap)[tableName]

	// 获取当前表字段注释
	/*	columnCommentMap, err := this.getTableColumnComment(tableName)
		if err != nil {
			return nil, err
		}*/

	// 获取表的列信息
	/*columnTypes, err := migrator.ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}*/

	// 找到当前表所有字段
	columnList, ok := (*tableColumnInfoMap)[tableName]
	if !ok {
		return nil, errors.New("未找到表字段信息")
	}

	// 创建表信息
	tableInfo := &models.TableInfo{
		DatabaseName: databaseName,
		TableName:    tableName,
		ColumnList:   columnList,
		TableType:    tableType,
		Comment:      tableCommemt,
		IndexList:    indexInfoList,
	}

	return tableInfo, nil
}

// GetDatabaseInfo 生成数据库信息
func (this *DbDictService) GetDatabaseInfo(dbConfig *configs.DatabaseConfig) (*models.DatabaseInfo, error) {
	// 获取migrator对象
	migrator := this.DB.Migrator()

	// 获取数据库名称（schema）
	databaseName := migrator.CurrentDatabase()
	// 获取全库对象类型（按表聚合）
	tableTypeMap, err := this.getTableType(dbConfig)
	if err != nil {
		return nil, err
	}
	// 获取全库索引（按表聚合）
	indexInfoListMap, err := this.getTableIndexInfoMap(dbConfig)
	if err != nil {
		return nil, err
	}
	// 获取全库字段注释（按表聚合）
	tableCommentMap, err := this.getTableComment(dbConfig)
	if err != nil {
		return nil, err
	}
	// 获取全库字段类型（按表聚合）
	tableColumnInfoMap, err := this.getTableColumnInfoMap(dbConfig)
	if err != nil {
		return nil, err
	}

	// 获取所有表名
	tableList := make([]string, 0, len(tableTypeMap))
	for tableName := range tableTypeMap {
		tableList = append(tableList, tableName)
	}
	// 排序
	sort.Strings(tableList)
	if err != nil {
		return nil, err
	}

	// 结果
	tableNameList := []string{}
	tableMap := make(map[string]models.TableInfo)
	// 遍历表
	for _, tableName := range tableList {
		tableInfo, err := this.buildTableInfo(databaseName, tableName, &tableTypeMap, &tableColumnInfoMap, &indexInfoListMap, &tableCommentMap)
		if err != nil {
			continue
		}

		// 添加到结果
		tableMap[tableName] = *tableInfo
		tableNameList = append(tableNameList, tableName)
	}
	// 排序
	sort.Strings(tableNameList)

	// 生成数据库信息
	dbInfo := &models.DatabaseInfo{
		DatabaseName:  this.DB.Migrator().CurrentDatabase(),
		TableCount:    len(tableList),
		TableNameList: tableNameList,
		TableMap:      tableMap,
	}

	return dbInfo, nil
}

//  ----- BUILD --------------------

// BuildAll 生成数据库
func (this *DbDictService) BuildAll(dbConfig *configs.DatabaseConfig, outputDirPath string, format string, overwrite bool) (result []string, err error) {
	/* 获取数据库信息 */
	databaseInfo, err := this.GetDatabaseInfo(dbConfig)
	if nil != err {
		slog.Error("生成数据库字典失败", "error", err)
		return nil, err
	}
	slog.Debug("数据库字典", "databaseInfo", databaseInfo)

	// 数据表数量（包含索引页）
	total := len(databaseInfo.TableMap) + 1
	// 当前数量
	current := 0

	/* 生成数据库信息 */
	// 计数
	current++
	cur, err := this.rendering(dbConfig, format, databaseInfo, outputDirPath, overwrite, total, current)
	if nil != err {
		return nil, err
	}
	// 如果结果中不包含，则添加到结果
	if !slices.Contains(result, cur) {
		result = append(result, cur)
	}

	/* 生成数据表信息 */
	// 排序过的表名
	tableNameList := databaseInfo.TableNameList
	// 数据表map
	tableInfoMap := databaseInfo.TableMap
	// 遍历保存
	for _, tableName := range tableNameList {
		// 计数
		current++
		// 当前表信息
		tableInfo := tableInfoMap[tableName]
		// 保存到md文件
		cur, err := this.rendering(dbConfig, format, &tableInfo, outputDirPath, overwrite, total, current)
		if nil != err {
			continue
		}

		// 如果结果中不包含，则添加到结果
		if !slices.Contains(result, cur) {
			result = append(result, cur)
		}
	}

	return result, nil
}
