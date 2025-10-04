package services

var (
	sql_getTableIndexInfoMap = map[string]string{
		// SQL Server 查询表类型
		"sqlserver": `
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
		`,
		// MySQL 查询表类型
		"mysql": `
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
		`,
		// PostgresSQL 查询表类型
		"postgres": `
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
		`,
		// Oracle 查询对象类型
		"oracle": `
			SELECT
				SYS_CONTEXT('USERENV', 'DB_NAME') AS "database_name"
			  , ai.owner AS "schema_name"
			  , ai.table_name AS "table_name"
			  , ai.index_name AS "index_name"
			  , ai.index_type AS "index_type"
			  , LISTAGG(aic.column_name, ', ') WITHIN GROUP (ORDER BY aic.column_position) AS "column_names"
			  , CASE
					WHEN ai.uniqueness = 'UNIQUE' THEN 1
					ELSE 0
					END AS "is_unique"
			  , CASE
					WHEN EXISTS (
						SELECT
							1
						FROM all_constraints ac
						WHERE
							  ac.owner = ai.owner
						  AND ac.table_name = ai.table_name
						  AND ac.constraint_type = 'P'
						  AND ac.constraint_name = ai.index_name
					) THEN 1
					ELSE 0
					END AS "is_primary"
			  , NULL AS "index_comment"
			FROM all_indexes ai
				 JOIN all_ind_columns aic
				 ON ai.index_name = aic.index_name
					 AND ai.table_name = aic.table_name
					 AND ai.owner = aic.index_owner
			WHERE
				ai.owner = UPPER(?)
			GROUP BY
				SYS_CONTEXT('USERENV', 'DB_NAME')
			  , ai.owner
			  , ai.table_name
			  , ai.index_name
			  , ai.index_type
			  , ai.uniqueness
			ORDER BY
				"database_name"
			  , "schema_name"
			  , "table_name"
			  , "index_name"
			  , "index_type"
		`,
		// SQLite 查询表类型
		"sqlite": `
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
		`,
	}
)
