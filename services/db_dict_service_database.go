package services

import (
	"errors"
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

func (this *DbDictService) getTableColumnInfoMap(databaseName string) (map[string][]*models.ColumnInfo, error) {
	// 数据库类型
	dbType := this.DB.Dialector.Name()

	// 不同类型不同处理方法
	dataList := []*models.ColumnInfo{}

	// 查询
	var query, ok = sql_getTableColumnInfos_map[dbType]
	if !ok {
		return nil, errors.New("不支持的数据库类型")
	}

	// 参数
	params := []interface{}{databaseName}

	// 调用
	err := this.DB.Raw(query, params...).Scan(&dataList).Error
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
func (this *DbDictService) getTableIndexInfoMap(databaseName string) (map[string][]*models.IndexInfo, error) {
	// 数据库类型
	dbType := this.DB.Dialector.Name()

	// 不同类型不同处理方法
	dataList := []*models.IndexInfo{}
	// SQL
	var query string

	switch dbType {
	case "sqlserver":
		// SQL Server 查询索引信息
		query = `
			SELECT
				DB_NAME() AS database_name
			  , sc.name AS schema_name
			  , t.name AS table_name
			  , i.name AS index_name
			  , i.type_desc AS index_type
			  , i.is_primary_key AS is_primary
			  , i.is_unique AS is_unique
			  , ep.[value] AS index_comment
			  , STUFF((
						  SELECT
							  ',' + col.name + ' ' +
							  CASE WHEN ic.is_descending_key = 1 THEN 'DESC' ELSE 'ASC' END
						  FROM sys.index_columns ic
							   INNER JOIN sys.columns col
							   ON ic.object_id = col.object_id
								   AND ic.column_id = col.column_id
						  WHERE
								ic.object_id = i.object_id
							AND ic.index_id = i.index_id
						  ORDER BY
							  ic.key_ordinal
						  FOR XML PATH('')
					  ), 1, 1, '') AS column_names
			FROM sys.tables t
				 LEFT JOIN sys.schemas sc
				 ON t.schema_id = sc.schema_id
				 LEFT JOIN sys.indexes i
				 ON t.object_id = i.object_id
				 LEFT JOIN sys.extended_properties ep
				 ON ep.major_id = t.object_id -- major_id = 表的ID（索引所属表）
					 AND ep.minor_id = i.index_id -- minor_id = 索引的ID（区分同一表的不同索引）
					 AND ep.class = 7 -- class=7 表示「索引」类型（固定值）
					 AND ep.name = 'MS_Description' -- 注释的属性名（默认用MS_Description存储注释）
			WHERE
				  t.type = 'U' -- t.type='U' 表示仅用户表（排除系统表）
			  AND i.type <> 0 -- i.type<>0 表示排除无效索引（0=堆，无索引）
			  AND DB_NAME() = ?
			ORDER BY
				database_name
			  , schema_name
			  , table_name
			  , index_name
		`
	case "mysql":
		// MySQL 查询索引信息
		query = `
			SELECT
				DATABASE() AS database_name
			  , t.TABLE_CATALOG AS catalog_name
			  , t.TABLE_SCHEMA AS schema_name
			  , t.TABLE_NAME AS table_name
			  , t.INDEX_NAME AS index_name
			  , t.INDEX_TYPE AS index_type
			  , GROUP_CONCAT(t.COLUMN_NAME ORDER BY SEQ_IN_INDEX) AS column_names
			  , CASE WHEN t.NON_UNIQUE = 0 THEN TRUE ELSE FALSE END AS is_unique
			  , CASE WHEN t.INDEX_NAME = 'PRIMARY' THEN TRUE ELSE FALSE END AS is_primary
			  , t.INDEX_COMMENT AS index_comment
			FROM INFORMATION_SCHEMA.STATISTICS t
			WHERE t.TABLE_SCHEMA = ?
			GROUP BY
				t.INDEX_NAME
			  , t.NON_UNIQUE
			  , t.INDEX_TYPE
			ORDER BY
				database_name
			  , schema_name
			  , table_name
			  , index_type
		`
	case "postgres":
		// PostgreSQL 查询索引信息
		query = `
			SELECT
				CURRENT_DATABASE() AS database_name
			  , n.nspname AS schema_name
			  , t.relname AS table_name
			  , i.relname AS index_name
			  , am.amname AS index_type
			  , array_to_string(array_agg(a.attname ORDER BY array_position(idx.indkey, a.attnum)), ', ') AS column_names
			  , idx.indisunique AS is_unique
			  , idx.indisprimary AS is_primary
			  , (
					SELECT
						description
					FROM pg_catalog.pg_description
					WHERE
						objoid = i.oid
					  AND objsubid = 0
				) AS index_comment -- 使用子查询获取索引注释
			FROM pg_class t
				 JOIN pg_namespace n
				 ON t.relnamespace = n.oid
				 JOIN pg_index idx
				 ON t.oid = idx.indrelid
				 JOIN pg_class i
				 ON idx.indexrelid = i.oid
				 JOIN pg_am am
				 ON i.relam = am.oid
				 JOIN pg_attribute a
				 ON a.attrelid = t.oid AND a.attnum = ANY (idx.indkey
				 )
			WHERE
				  t.relkind = 'r'
			  AND n.nspname NOT IN ('pg_catalog', 'information_schema')
			  AND current_database() = ?
			GROUP BY
				n.nspname
			  , t.relname
			  , i.relname
			  , i.oid
			  , am.amname
			  , idx.indisunique
			  , idx.indisprimary
			ORDER BY
				database_name
			  , schema_name
			  , table_name
			  , index_name
		`
	case "oracle":
		// Oracle 查询索引信息
		query = `
			SELECT 
				SYS_CONTEXT('USERENV', 'DB_NAME') AS database_name,
				NULL AS schema_name,  -- Oracle中通常用USER表示当前模式，如需获取其他模式需调整
				ui.table_name AS table_name, 
				ui.index_name AS index_name,
				ui.index_type AS index_type,
				LISTAGG(uic.column_name, ', ') WITHIN GROUP (ORDER BY uic.column_position) AS column_names,
				CASE WHEN ui.uniqueness = 'UNIQUE' THEN TRUE ELSE FALSE END AS is_unique,
				CASE WHEN EXISTS (
					SELECT 1 FROM user_constraints uc 
					WHERE uc.table_name = ui.table_name 
					AND uc.constraint_type = 'P' 
					AND uc.constraint_name = ui.index_name
				) THEN TRUE ELSE FALSE END AS is_primary,
				NULL AS index_comment  -- Oracle系统视图中通常不直接提供索引注释，需另寻方法
			FROM user_indexes ui
			JOIN user_ind_columns uic ON ui.index_name = uic.index_name AND ui.table_name = uic.table_name
			WHERE SYS_CONTEXT('USERENV', 'DB_NAME') = ?
			GROUP BY 
				SYS_CONTEXT('USERENV', 'DB_NAME'),
				ui.table_name,
				ui.index_name, 
				ui.uniqueness, 
				ui.index_type
			ORDER BY 
				database_name, 
				schema_name, 
				table_name, 
				index_type;
		`
	case "sqlite":
		// SQLite 查询索引信息
		query = `
			SELECT
				'main' AS database_name
			  , NULL AS schema_name
			  , t.tbl_name AS table_name
			  , t.name AS index_name
			  , NULL AS index_type
			  , (
					SELECT
						TRIM(REPLACE(REPLACE(SUBSTR(REPLACE(REPLACE(tt.sql, CHAR(13), ''), CHAR(10), ''), INSTR(REPLACE(REPLACE(tt.sql, CHAR(13), ''), CHAR(10), ''), '(') + 1, (INSTR(REPLACE(REPLACE(tt.sql, CHAR(13), ''), CHAR(10), ''), ')') - INSTR(REPLACE(REPLACE(tt.sql, CHAR(13), ''), CHAR(10), ''), '(')) - 1), '"', ''), ',  ', ',')) AS column_names
					FROM sqlite_master tt
					WHERE
						  tt.type = 'index'
					  AND t.name = tt.name
				) AS column_names
			  , (
					SELECT u.[unique]
					FROM PRAGMA_INDEX_LIST(t.tbl_name) u
					WHERE t.name = u.name
				) AS is_unique
			  , (
					SELECT u.origin = 'pk'
					FROM PRAGMA_INDEX_LIST(t.tbl_name) u
					WHERE t.name = u.name
				) AS is_primary
			  , NULL AS index_comment
			FROM sqlite_master t
			WHERE
				type = 'index'
			AND database_name = ?
			ORDER BY
				database_name
			  , schema_name
			  , table_name
			  , index_type
		`
	default:
		// 不支持的数据库类型
	}

	// 参数
	params := []interface{}{databaseName}

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
func (this *DbDictService) getTableComment(databaseName string) (map[string]string, error) {
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
		params := []interface{}{databaseName}

		// 执行
		err := this.DB.Raw(query, params...).Scan(&tableComments).Error
		if err != nil {
			return nil, err
		}
	case "postgres":
		// PostgreSQL 从 information_schema.tables 获取表注释
		query := `
			SELECT 
				table_name AS table_name,
				obj_description((quote_ident(table_schema) || '.' || quote_ident(table_name))::regclass, 'pg_class') AS comment
			FROM information_schema.tables 
			WHERE 
				table_type IN ('BASE TABLE', 'VIEW')
-- 				table_catalog = ?
				AND table_schema = ?
        `
		// PostgresSQL需要提供databaseName
		params := []interface{}{databaseName}

		// 执行
		err := this.DB.Raw(query, params...).Scan(&tableComments).Error
		if err != nil {
			return nil, err
		}
	case "oracle":
		// Oracle 从 ALL_TAB_COMMENTS 获取表注释
		query := `
			SELECT 
				t.table_name AS table_name,
				pg_catalog.obj_description(c.oid, 'pg_class') AS comment
			FROM information_schema.tables t
			JOIN pg_catalog.pg_class c ON c.relname = t.table_name
			JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.table_schema
			WHERE 
				t.table_type IN ('BASE TABLE', 'VIEW')
				AND t.table_schema = ?
        `
		// Oracle不需要指定表名
		params := []interface{}{databaseName}

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
		// PostgreSQL 从 information_schema 获取列注释
		query := `
            SELECT 
                column_name,
                col_description(
                    (quote_ident(table_schema)||'.'||quote_ident(table_name))::regclass::oid, 
                    ordinal_position
                ) AS comment
            FROM information_schema.columns 
            WHERE table_name = $1 
            AND table_schema = current_schema()
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
func (this *DbDictService) getTableType(databaseName string) (map[string]string, error) {
	var dataList []*models.TableType

	dbType := this.DB.Dialector.Name()
	if "sqlserver" == dbType {
		// SQL Server 查询表类型
		query := `
            SELECT
				o.name AS table_name
			  , (CASE o.[type_desc]
					 WHEN 'VIEW' THEN 'view'
					 WHEN 'USER_TABLE' THEN 'table'
					 ELSE 'other'
				END) AS table_type
			FROM SYS.OBJECTS o
			WHERE
				TYPE IN ('U', 'V')
            	AND DB_NAME() = ?
			ORDER BY
			    table_name
			  , table_type
        `
		err := this.DB.Raw(query, databaseName).Scan(&dataList).Error
		if err != nil {
			return nil, err
		}
	} else if "mysql" == dbType {
		// MySQL 查询表类型
		query := `
            SELECT
				t.TABLE_NAME AS table_name
			  , (CASE t.TABLE_TYPE
					 WHEN 'BASE TABLE' THEN 'table'
					 WHEN 'VIEW' THEN 'view'
					 WHEN 'SYSTEM VIEW' THEN 'sys_view'
					 ELSE 'Other'
				END) AS table_type
			FROM INFORMATION_SCHEMA.TABLES t
			WHERE
				t.TABLE_SCHEMA = ?
			ORDER BY
			    table_name
			  , table_type
        `
		err := this.DB.Raw(query, databaseName).Scan(&dataList).Error
		if err != nil {
			return nil, err
		}
	} else if "postgres" == dbType {
		// PostgreSQL 查询表类型
		query := `
            SELECT
				  t.table_name AS table_name
				, (CASE t.table_type
					 WHEN 'BASE TABLE' THEN 'table'
					 WHEN 'VIEW' THEN 'view'
					 ELSE 'Other'
				END) AS table_type
			FROM information_schema.tables t
			WHERE
			      t.table_schema NOT IN ('pg_catalog', 'information_schema')  
				AND t.table_schema = ?  
			ORDER BY
			    table_name
			  , table_type
        `
		err := this.DB.Raw(query, databaseName).Scan(&dataList).Error
		if err != nil {
			return nil, err
		}
	} else if "oracle" == dbType {
		// Oracle 查询对象类型
		query := `
			SELECT
				(SELECT GLOBAL_NAME FROM GLOBAL_NAME) AS DatabaseName,
				t.OBJECT_NAME AS table_name,
				(CASE t.OBJECT_TYPE
					 WHEN 'TABLE' THEN 'table'
					 WHEN 'VIEW' THEN 'view'
					 ELSE 'Other'
				END) AS table_type
			FROM ALL_OBJECTS t
			WHERE 
				t.OWNER = ?  
				AND t.OBJECT_TYPE IN ('TABLE', 'VIEW')
			ORDER BY
			    table_name
			  , table_type
	    `
		err := this.DB.Raw(query, databaseName).Scan(&dataList).Error
		if err != nil {
			return nil, err
		}
	} else if "sqlite" == dbType {
		// SQLite 查询表类型
		query := `
            SELECT
				t.name AS table_name
			  , (CASE t.[type]
					 WHEN 'table' THEN 'table'
					 WHEN 'view' THEN 'view'
				END) AS table_type
			FROM sqlite_master t
			WHERE
				  t.type IN ('table', 'view')
			  AND DatabaseName = 'main'
			ORDER BY
			    table_name
			  , table_type
        `
		err := this.DB.Raw(query, databaseName).Scan(&dataList).Error
		if err != nil {
			return nil, err
		}
	}

	// 转换为map
	result := make(map[string]string)
	for _, item := range dataList {
		result[item.TableName] = item.TableType
	}

	return result, nil
}
