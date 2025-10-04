package services

import (
	"errors"
	"goDict/configs"
	"goDict/models"
)

// TableComment 表注释信息
type TableComment struct {
	TableName string `json:"table_name"`
	Comment   string `json:"comment"`
}

// ColumnComment 列注释信息
type ColumnComment struct {
	ColumnName string `json:"column_name"`
	Comment    string `json:"comment"`
}

func (this *DbDictService) getTableColumnInfoMap(dbConfig *configs.DatabaseConfig) (map[string][]*models.ColumnInfo, error) {
	// 数据库类型
	dbType := this.DB.Dialector.Name()

	// 不同类型不同处理方法
	dataList := []*models.ColumnInfo{}

	// 查询
	var query, ok = sql_getTableColumnInfosMap[dbType]
	if !ok {
		return nil, errors.New("不支持的数据库类型")
	}

	var err error
	var params []interface{}
	if "oracle" == dbType {
		// 参数
		params = []interface{}{dbConfig.Database, dbConfig.Database, dbConfig.Database}
	} else {
		// 参数
		params = []interface{}{dbConfig.Database}
	}

	// 调用
	err = this.DB.Raw(query, params...).Scan(&dataList).Error
	if err != nil {
		return nil, err
	}

	// 将数据根据tableName聚合
	result := make(map[string][]*models.ColumnInfo, len(dataList))
	for _, columnInfo := range dataList {
		// 提取表名
		tableName := columnInfo.TableName

		// 根据表名找到map，如果不存在先创建
		columnInfoList, ok := result[tableName]
		if !ok {
			columnInfoList = []*models.ColumnInfo{}
		}

		// 保存数据库索引到列表
		columnInfoList = append(columnInfoList, columnInfo)

		// 更新结果
		result[tableName] = columnInfoList
	}

	return result, nil
}
func (this *DbDictService) getTableIndexInfoMap(dbConfig *configs.DatabaseConfig) (map[string][]*models.IndexInfo, error) {
	// 数据库类型
	dbType := this.DB.Dialector.Name()

	// 不同类型不同处理方法
	dataList := []*models.IndexInfo{}
	// SQL
	query := sql_getTableIndexInfoMap[dbType]
	// 参数
	params := []interface{}{dbConfig.Database}
	// 调用
	err := this.DB.Raw(query, params...).Scan(&dataList).Error
	if err != nil {
		return nil, err
	}

	// 将数据根据tableName聚合
	result := make(map[string][]*models.IndexInfo, len(dataList))
	for _, indexInfo := range dataList {
		// 提取表名
		tableName := indexInfo.TableName

		// 根据表名找到map，如果不存在先创建
		indexList, ok := result[tableName]
		if !ok {
			indexList = []*models.IndexInfo{}
		}

		// 保存数据库索引到列表
		indexList = append(indexList, indexInfo)

		// 更新结果
		result[tableName] = indexList
	}

	return result, nil
}

// getTableComment 表注释信息
func (this *DbDictService) getTableComment(dbConfig *configs.DatabaseConfig) (map[string]string, error) {
	var tableComments []TableComment

	dbType := this.DB.Dialector.Name()

	switch dbType {
	case "sqlserver":
		// SQL Server 使用 sys.tables 和 sys.extended_properties 获取表注释
		query := `
			SELECT
				t.name AS table_name
			  , CAST(ISNULL(ep.value, '') AS NVARCHAR(4000)) AS comment
			FROM sys.tables t
				 LEFT JOIN sys.extended_properties ep
				 ON ep.major_id = t.object_id
					 AND ep.minor_id = 0
					 AND ep.name = 'MS_Description'
			WHERE
				t.is_ms_shipped = 0
        `
		// SQLServer不需要提供databaseName
		params := []interface{}{}

		// 执行
		err := this.DB.Raw(query, params...).Scan(&tableComments).Error
		if err != nil {
			return nil, err
		}
	case "mysql":
		// MySQL 从 information_schema.tables 获取表注释
		query := `
			SELECT
				TABLE_NAME AS table_name
			  , TABLE_COMMENT AS comment
			FROM INFORMATION_SCHEMA.TABLES
			WHERE
				  TABLE_TYPE IN ('VIEW', 'BASE TABLE')
			  AND TABLE_SCHEMA = ? 
        `
		// MySQL需要提供databaseName
		params := []interface{}{dbConfig.Database}

		// 执行
		err := this.DB.Raw(query, params...).Scan(&tableComments).Error
		if err != nil {
			return nil, err
		}
	case "postgres":
		// PostgresSQL 从 information_schema.tables 获取表注释
		query := `
			SELECT 
				table_name AS table_name,
				obj_description((quote_ident(table_schema) || '.' || quote_ident(table_name))::regclass, 'pg_class') AS comment
			FROM information_schema.tables 
			WHERE 
				table_type IN ('BASE TABLE', 'VIEW')
				AND table_catalog = ?
        `
		// PostgresSQL需要提供databaseName
		params := []interface{}{dbConfig.Database}

		// 执行
		err := this.DB.Raw(query, params...).Scan(&tableComments).Error
		if err != nil {
			return nil, err
		}
	case "oracle":
		// Oracle 从 ALL_TAB_COMMENTS 获取表注释
		query := `
			SELECT
				TABLE_NAME AS "table_name"
			  , COMMENTS AS "comment"
			FROM ALL_TAB_COMMENTS
			WHERE
				OWNER = UPPER(?)
			ORDER BY TABLE_NAME
        `
		// Oracle需要提供databaseName
		params := []interface{}{dbConfig.Database}

		// 执行
		err := this.DB.Raw(query, params...).Scan(&tableComments).Error
		if err != nil {
			return nil, err
		}
	case "sqlite":
		// SQLite 不直接支持表注释，返回空列表
		tableComments = []TableComment{}
	default:
		// 其他数据库类型也返回空列表
		tableComments = []TableComment{}
	}

	// 转换为map
	result := make(map[string]string)
	for _, comment := range tableComments {
		result[comment.TableName] = comment.Comment
	}

	return result, nil
}

// getColumnCommentByTableName 获取表中各列的注释信息（主要用于SQL Server）
func (this *DbDictService) getTableColumnComment(tableName string) (map[string]string, error) {
	var columnComments []ColumnComment

	dbType := this.DB.Dialector.Name()

	switch dbType {
	case "sqlserver":
		// SQL Server 使用 fn_listextendedproperty 函数获取列注释
		query := `
            SELECT
				c.name AS column_name,
				CAST(ISNULL(ep.value, '') AS NVARCHAR(4000)) AS comment -- 将NULL转换为空字符串
			FROM sys.columns c
				 LEFT JOIN sys.extended_properties ep ON ep.major_id = c.object_id
					AND ep.minor_id = c.column_id
					AND ep.name = 'MS_Description'
			WHERE c.object_id = OBJECT_ID(?)
        `
		err := this.DB.Raw(query, tableName).Scan(&columnComments).Error
		if err != nil {
			return nil, err
		}
	case "mysql":
		// MySQL 从 information_schema 获取列注释
		query := `
            SELECT 
                COLUMN_NAME AS column_name,
                COLUMN_COMMENT AS comment
            FROM INFORMATION_SCHEMA.COLUMNS 
            WHERE TABLE_NAME = ? 
            AND TABLE_SCHEMA = DATABASE()
            AND COLUMN_COMMENT != ''
        `
		err := this.DB.Raw(query, tableName).Scan(&columnComments).Error
		if err != nil {
			return nil, err
		}
	case "postgres":
		// PostgresSQL 从 information_schema 获取列注释
		query := `
            SELECT 
                column_name,
                col_description(
                    (quote_ident(table_schema)||'.'||quote_ident(table_name))::regclass::oid, 
                    ordinal_position
                ) AS comment
            FROM information_schema.columns 
            WHERE table_catalog = current_catalog
				AND table_name = ?
				AND col_description(
					(quote_ident(table_schema)||'.'||quote_ident(table_name))::regclass::oid, 
					ordinal_position
            ) IS NOT NULL
        `
		err := this.DB.Raw(query, tableName).Scan(&columnComments).Error
		if err != nil {
			return nil, err
		}
	case "oracle":
		// Oracle 从 ALL_COL_COMMENTS 获取列注释
		query := `
            SELECT 
                COLUMN_NAME AS column_name,
                COMMENTS AS comment
            FROM ALL_COL_COMMENTS 
            WHERE TABLE_NAME = UPPER(:1) 
            AND OWNER = (SELECT USER FROM DUAL)
            AND COMMENTS IS NOT NULL
        `
		err := this.DB.Raw(query, tableName).Scan(&columnComments).Error
		if err != nil {
			return nil, err
		}
	case "sqlite":
		// SQLite 不直接支持列注释，返回空列表
		columnComments = []ColumnComment{}

	default:
		// 其他数据库类型也返回空列表
		columnComments = []ColumnComment{}
	}

	// 转换为map
	result := make(map[string]string)
	for _, comment := range columnComments {
		result[comment.ColumnName] = comment.Comment
	}

	return result, nil
}

// getTableType 获取表类型（兼容SQL Server）
func (this *DbDictService) getTableType(dbConfig *configs.DatabaseConfig) (map[string]string, error) {
	var dataList []*models.TableType

	// 类型
	dbType := this.DB.Dialector.Name()
	// SQL
	query := sql_getTableTypeMap[dbType]
	// 参数
	var params []interface{}
	// Sqlite不需要传递参数，其他都需要传递
	if "sqlite" != dbType {
		params = append(params, dbConfig.Database)
	}
	// 执行
	err := this.DB.Raw(query, params...).Scan(&dataList).Error
	if err != nil {
		return nil, err
	}

	// 转换为map
	result := make(map[string]string)
	for _, item := range dataList {
		result[item.TableName] = item.TableType
	}

	return result, nil
}
