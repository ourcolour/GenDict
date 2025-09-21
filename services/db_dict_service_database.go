package services

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

// getTableComment 表注释信息
func (this *DbDictService) getTableComment(databaseName string, tableName *string) (map[string]string, error) {
	var tableComments []TableComment

	dbType := this.DB.Dialector.Name()

	switch dbType {
	case "sqlserver":
		// SQL Server 使用 sys.tables 和 sys.extended_properties 获取表注释
		query := `
            SELECT 
                t.name AS table_name,
                CAST(ISNULL(ep.value, '') AS NVARCHAR(4000)) AS comment
            FROM sys.tables t
            LEFT JOIN sys.extended_properties ep ON ep.major_id = t.object_id
                AND ep.minor_id = 0
                AND ep.name = 'MS_Description'
            WHERE t.is_ms_shipped = 0
        `
		if nil != tableName {
			query += " AND t.name = ?"
		}
		err := this.DB.Raw(query, *tableName).Scan(&tableComments).Error
		if err != nil {
			return nil, err
		}
	case "mysql":
		// MySQL 从 information_schema.tables 获取表注释
		query := `
            SELECT 
                TABLE_NAME AS table_name,
                TABLE_COMMENT AS comment
            FROM INFORMATION_SCHEMA.TABLES 
            WHERE TABLE_SCHEMA = ?
            AND TABLE_TYPE = 'BASE TABLE'
            AND TABLE_COMMENT != ''
        `
		if nil != tableName {
			query += " AND TABLE_NAME = ?"
		}
		err := this.DB.Raw(query, databaseName, *tableName).Scan(&tableComments).Error
		if err != nil {
			return nil, err
		}
	case "postgres":
		// PostgreSQL 从 information_schema.tables 获取表注释
		query := `
            SELECT 
                table_name,
                obj_description((quote_ident(table_schema)||'.'||quote_ident(table_name))::regclass) AS comment
            FROM information_schema.tables 
            WHERE table_catalog = ?
            AND table_schema = 'public'
            AND table_type = 'BASE TABLE'
            AND obj_description((quote_ident(table_schema)||'.'||quote_ident(table_name))::regclass) IS NOT NULL
        `
		if nil != tableName {
			query += " AND table_name = ?"
		}
		err := this.DB.Raw(query, databaseName, *tableName).Scan(&tableComments).Error
		if err != nil {
			return nil, err
		}
	case "oracle":
		// Oracle 从 ALL_TAB_COMMENTS 获取表注释
		query := `
            SELECT 
                TABLE_NAME AS table_name,
                COMMENTS AS comment
            FROM ALL_TAB_COMMENTS 
            WHERE OWNER = UPPER(:1)
            AND TABLE_TYPE = 'TABLE'
            AND COMMENTS IS NOT NULL
        `
		if nil != tableName {
			query += " AND TABLE_NAME = ?"
		}
		err := this.DB.Raw(query, databaseName, *tableName).Scan(&tableComments).Error
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
func (this *DbDictService) getTableType(tableName string) (string, error) {
	var tableType string

	dbType := this.DB.Dialector.Name()
	if "sqlserver" == dbType {
		// SQL Server 查询表类型
		query := `
            SELECT TYPE_DESC 
            FROM SYS.OBJECTS 
            WHERE NAME = ? AND TYPE IN ('U', 'V')
        `
		err := this.DB.Raw(query, tableName).Scan(&tableType).Error
		if err != nil {
			return "", err
		}

		// 根据返回值判断类型
		if tableType == "USER_TABLE" {
			return "table", nil
		} else if tableType == "VIEW" {
			return "view", nil
		}
	} else if "mysql" == dbType {
		// MySQL 查询表类型
		query := `
            SELECT TABLE_TYPE 
            FROM INFORMATION_SCHEMA.TABLES 
            WHERE TABLE_NAME = ? AND TABLE_SCHEMA = DATABASE()
        `
		err := this.DB.Raw(query, tableName).Scan(&tableType).Error
		if err != nil {
			return "", err
		}

		if tableType == "VIEW" {
			return "view", nil
		}
		return "table", nil
	} else if "postgres" == dbType {
		// PostgreSQL 查询表类型
		query := `
            SELECT table_type 
            FROM information_schema.tables 
            WHERE table_name = $1 AND table_schema = current_schema()
        `
		err := this.DB.Raw(query, tableName).Scan(&tableType).Error
		if err != nil {
			return "", err
		}

		if tableType == "VIEW" {
			return "view", nil
		}
		return "table", nil
	} else if "oracle" == dbType {
		// Oracle 查询对象类型
		query := `
			SELECT OBJECT_TYPE 
			FROM ALL_OBJECTS 
			WHERE OBJECT_NAME = UPPER(:1) 
			AND OWNER = (SELECT USER FROM DUAL) 
			AND OBJECT_TYPE IN ('TABLE', 'VIEW')
	    `
		err := this.DB.Raw(query, tableName).Scan(&tableType).Error
		if err != nil {
			return "", err
		}

		// 根据返回值判断类型 (Oracle 返回 'TABLE' 或 'VIEW')
		if tableType == "TABLE" {
			return "table", nil
		} else if tableType == "VIEW" {
			return "view", nil
		}
	} else if "sqlite" == dbType {
		// SQLite 查询表类型
		query := `
            SELECT type 
            FROM sqlite_master 
            WHERE name = ?
        `
		err := this.DB.Raw(query, tableName).Scan(&tableType).Error
		if err != nil {
			return "", err
		}

		if tableType == "view" {
			return "view", nil
		}
		return "table", nil
	}

	// 尝试使用默认的 TableType 方法
	tableTypeInfo, err := this.DB.Migrator().TableType(tableName)
	if err != nil {
		return "", err
	}
	if "VIEW" == tableTypeInfo.Type() {
		return "view", nil
	}
	return "table", nil
}
