package services

import (
	"errors"
	"fmt"
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

// getTableInfo 获取表信息
func (this *DbDictService) getTableInfo(tableName string) (*models.TableInfo, error) {
	migrator := this.DB.Migrator()

	// 获取数据库名称（schema）
	databaseName := migrator.CurrentDatabase()
	// 确定当前对象类型（table / view）
	tableType, err := this.getTableType(tableName)
	if err != nil {
		return nil, err
	}

	// 如果为空，尝试从查询结果中获取字段注释
	tableCommentMap, err := this.getTableComment(databaseName, &tableName)
	if err != nil {
		return nil, err
	}

	// 获取表的列信息
	columnTypes, err := migrator.ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}

	// 获取当前表字段注释
	columnCommentMap, err := this.getTableColumnComment(tableName)
	if err != nil {
		return nil, err
	}

	tableInfo := &models.TableInfo{
		DatabaseName: databaseName,
		TableName:    tableName,
		ColumnList:   make([]models.ColumnInfo, 0, len(columnTypes)),
		TableType:    tableType,
		Comment:      tableCommentMap[tableName],
	}

	for _, columnType := range columnTypes {
		// 获取列名
		name := columnType.Name()

		// 获取数据类型
		dbType := columnType.DatabaseTypeName()

		// 获取是否可为NULL
		nullable, ok := columnType.Nullable()
		if !ok {
			nullable = true // 默认假设可为NULL
		}

		// 获取默认值
		defaultValue, ok := columnType.DefaultValue()
		if !ok {
			defaultValue = ""
		}

		// 获取注释
		comment, ok := columnType.Comment()
		if !ok {
			comment = ""
		}
		// 如果为空，尝试从查询结果中获取字段注释
		if comment == "" {
			comment = columnCommentMap[name]
		}

		// 检查是否是主键
		isPrimary, ok := columnType.PrimaryKey()
		if !ok {
			isPrimary = false
		}

		// 检查是否是自增
		isAutoIncrement, ok := columnType.AutoIncrement()
		if !ok {
			isAutoIncrement = false
		}

		// 检查是否是唯一
		isUnique, ok := columnType.Unique()
		if !ok {
			isUnique = false
		}

		// 检查是否是唯一
		length, ok := columnType.Length()
		if !ok {
			length = 0
		}

		// 检查数据精度
		var decimalSize models.DecimalSizeInfo
		precision, scale, ok := columnType.DecimalSize()
		if !ok {
			decimalSize = models.DecimalSizeInfo{
				Precision: 0,
				Scale:     0,
			}
		} else {
			decimalSize = models.DecimalSizeInfo{
				Precision: precision,
				Scale:     scale,
			}
		}

		tableInfo.ColumnList = append(tableInfo.ColumnList, models.ColumnInfo{
			Name:            name,
			Type:            dbType,
			Length:          length,
			Nullable:        nullable,
			Default:         defaultValue,
			Comment:         comment,
			IsPrimary:       isPrimary,
			IsAutoIncrement: isAutoIncrement,
			IsUnique:        isUnique,
			DecimalSize:     decimalSize,
		})
	}

	return tableInfo, nil
}

// generateDatabase 生成数据库信息
func (this *DbDictService) generateDatabase() (*models.DatabaseInfo, error) {
	tableList, err := this.DB.Migrator().GetTables()
	if err != nil {
		return nil, err
	}

	// 结果
	tableNameList := []string{}
	tableMap := make(map[string]models.TableInfo)
	// 遍历表
	for _, tableName := range tableList {
		tableInfo, err := this.getTableInfo(tableName)
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

// BuildTableByName 生成数据库字典（根据表名）
func (this *DbDictService) BuildTableByName(format string, tableName string, outputDirPath string, overwrite bool, total int, current int) (string, error) {
	// 生成数据库字典
	tableInfo, err := this.getTableInfo(tableName)
	if nil != err {
		slog.Error("生成数据库字典失败", "error", err)
		return "", err
	}
	slog.Debug("数据库字典", "tableInfo", tableInfo)

	return this.buildByDataInfo(format, tableInfo, outputDirPath, overwrite, total, current)
}

// buildByDataInfo 生成数据库字典（根据数据）
func (this *DbDictService) buildByDataInfo(format string, templateData interface{}, outputDirPath string, overwrite bool, total int, current int) (string, error) {
	// Args
	if nil == templateData {
		return "", errors.New("请指定数据")
	}
	if "" == outputDirPath {
		return "", fmt.Errorf("请指定输出目录")
	}
	if "" == format || !SUPPORTED_FORMAT[format] {
		return "", fmt.Errorf("不支持的格式")
	}

	/* 判断数据类型 */
	// 判断类型
	if dataInfo, ok := templateData.(string); ok {
		// 如果是字符串
		return this.BuildTableByName(format, dataInfo, outputDirPath, overwrite, total, current)
	}

	// 根据模板生成内容
	return this.rendering(format, templateData, outputDirPath, overwrite, total, current)
}

// BuildAll 生成数据库
func (this *DbDictService) BuildAll(outputDirPath string, format string, overwrite bool) (result []string, err error) {
	// Args
	/* 获取数据库信息 */
	databaseInfo, err := this.generateDatabase()
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
	cur, err := this.buildByDataInfo(format, databaseInfo, outputDirPath, overwrite, total, current)
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
		cur, err := this.buildByDataInfo(format, &tableInfo, outputDirPath, overwrite, total, current)
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
