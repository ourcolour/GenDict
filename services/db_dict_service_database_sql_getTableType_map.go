package services

var (
	sql_getTableTypeMap = map[string]string{
		// SQL Server 查询表类型
		"sqlserver": `
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
		`,
		// MySQL 查询表类型
		"mysql": `
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
        `,
		// PostgresSQL 查询表类型
		"postgres": `
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
				AND t.table_catalog = ?  
			ORDER BY
			    table_name
			  , table_type
        `,
		// Oracle 查询对象类型
		"oracle": `
			SELECT
				(SELECT GLOBAL_NAME FROM GLOBAL_NAME) AS "database_name",
				t.OBJECT_NAME AS "table_name",
				(CASE t.OBJECT_TYPE
					 WHEN 'TABLE' THEN 'table'
					 WHEN 'VIEW' THEN 'view'
					 ELSE 'Other'
					END) AS "table_type"
			FROM ALL_OBJECTS t
			WHERE
				  t.OWNER = UPPER(?)
			  AND t.OBJECT_TYPE IN ('TABLE', 'VIEW')
			ORDER BY
				"table_name"
			  , "table_type"
	    `,
		// SQLite 查询表类型
		"sqlite": `
            SELECT
				t.name AS table_name
			  , (CASE t.[type]
					 WHEN 'table' THEN 'table'
					 WHEN 'view' THEN 'view'
				END) AS table_type
			FROM sqlite_master t
			WHERE
				  t.type IN ('table', 'view')
			ORDER BY
			    table_name
			  , table_type
        `,
	}
)
