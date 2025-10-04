package services

var (
	sql_getTableColumnInfosMap = map[string]string{
		"sqlserver": `
			SELECT
				ty.ORDINAL_POSITION AS 'sort'
			  , c.TABLE_CATALOG AS 'database_name'
			  , c.TABLE_SCHEMA AS 'schema_name'
			  , c.TABLE_NAME AS 'table_name'
			  , c.COLUMN_NAME AS 'column_name'
			  , ty.DATA_TYPE AS 'data_type'
			  , c.CHARACTER_MAXIMUM_LENGTH AS 'length'
			  , c.NUMERIC_PRECISION AS 'precision'
			  , c.NUMERIC_PRECISION_RADIX AS 'radix'
			  , c.NUMERIC_SCALE AS 'scale'
			  , CASE c.IS_NULLABLE WHEN 'YES' THEN 1 ELSE 0 END AS 'nullable'
			  , CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END AS 'is_primary'
			  , CASE COLUMNPROPERTY(OBJECT_ID(c.TABLE_SCHEMA + '.' + c.TABLE_NAME), c.COLUMN_NAME, 'IsIdentity')
					WHEN 1 THEN 1
					ELSE 0 END AS 'is_auto_increment'
			  , CASE WHEN uni.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END AS 'is_unique'
			  , c.COLUMN_DEFAULT AS 'default'
			  , ep.value AS 'comment'
			FROM INFORMATION_SCHEMA.COLUMNS c
				 LEFT JOIN (
							   SELECT
								   TABLE_NAME
								 , COLUMN_NAME
							   FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
							   WHERE
								   CONSTRAINT_NAME LIKE 'PK_%'
						   ) pk
				 ON c.TABLE_NAME = pk.TABLE_NAME AND c.COLUMN_NAME = pk.COLUMN_NAME
				 LEFT JOIN (
							   -- 查询唯一约束的字段
							   SELECT
								   kcu.TABLE_NAME
								 , kcu.COLUMN_NAME
							   FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
									JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
									ON tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
							   WHERE
								   tc.CONSTRAINT_TYPE = 'UNIQUE'
						   ) uni
				 ON c.TABLE_NAME = uni.TABLE_NAME AND c.COLUMN_NAME = uni.COLUMN_NAME
				 LEFT JOIN INFORMATION_SCHEMA.COLUMNS ty
				 ON c.TABLE_NAME = ty.TABLE_NAME AND c.COLUMN_NAME = ty.COLUMN_NAME
				 LEFT JOIN sys.extended_properties ep -- 获取字段备注
				 ON ep.major_id = OBJECT_ID(c.TABLE_SCHEMA + '.' + c.TABLE_NAME)
					 AND ep.minor_id = (
						 SELECT
							 column_id
						 FROM sys.columns
						 WHERE
							   object_id = OBJECT_ID(c.TABLE_SCHEMA + '.' + c.TABLE_NAME)
						   AND name = c.COLUMN_NAME
					 )
					 AND ep.name = 'MS_Description' -- 通常备注使用这个名称
					 AND ep.class = 1 -- 表示对象是列
			WHERE
				c.TABLE_CATALOG = ?
			ORDER BY
				c.ORDINAL_POSITION
		 `,
		// MySQL
		"mysql": `
			SELECT
				c.ORDINAL_POSITION AS 'sort'
			  , c.TABLE_CATALOG AS 'database_name'
			  , c.TABLE_SCHEMA AS 'schema_name'
			  , c.TABLE_NAME AS 'table_name'
			  , c.COLUMN_NAME AS 'column_name'
			  , c.DATA_TYPE AS 'data_type'
			  , c.CHARACTER_MAXIMUM_LENGTH AS 'length'
			  , COALESCE(c.NUMERIC_PRECISION) AS 'precision'
			  , COALESCE(c.NUMERIC_SCALE) AS 'scale'
			  , NULL AS 'radix'
			  , CASE c.IS_NULLABLE WHEN 'YES' THEN 1 ELSE 0 END AS 'nullable'
			  , CASE WHEN kcu.CONSTRAINT_NAME = 'PRIMARY' THEN 1 ELSE 0 END AS 'is_primary'
			  , CASE WHEN c.EXTRA LIKE '%auto_increment%' THEN 1 ELSE 0 END AS 'is_auto_increment'
			  , CASE WHEN tc.CONSTRAINT_TYPE = 'UNIQUE' THEN 1 ELSE 0 END AS 'is_unique'
			  , c.COLUMN_DEFAULT AS 'default'
			  , c.COLUMN_COMMENT AS 'comment'
			FROM INFORMATION_SCHEMA.COLUMNS c
				 LEFT JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
				 ON c.TABLE_SCHEMA = kcu.TABLE_SCHEMA
					 AND c.TABLE_NAME = kcu.TABLE_NAME
					 AND c.COLUMN_NAME = kcu.COLUMN_NAME
					 AND kcu.CONSTRAINT_NAME = 'PRIMARY'
				 LEFT JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
				 ON kcu.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
					 AND kcu.TABLE_SCHEMA = tc.TABLE_SCHEMA
					 AND kcu.TABLE_NAME = tc.TABLE_NAME
					 AND tc.CONSTRAINT_TYPE = 'UNIQUE'
			WHERE
				c.TABLE_SCHEMA = ?
			ORDER BY
				c.TABLE_CATALOG
			  , c.TABLE_SCHEMA
			  , c.TABLE_NAME
			  , c.ORDINAL_POSITION
			  , c.COLUMN_NAME
		`,
		"postgres": `
			SELECT
				c.ordinal_position AS sort,
				c.table_catalog AS database_name,
				c.table_schema AS schema_name,
				c.table_name AS table_name,
				c.column_name AS column_name,
				c.data_type AS data_type,
				c.character_maximum_length AS length,
				c.numeric_precision AS precision,
				c.numeric_precision_radix AS radix,
				c.numeric_scale AS scale,
				CASE c.is_nullable WHEN 'YES' THEN 1 ELSE 0 END AS nullable,
				CASE WHEN pk.column_name IS NOT NULL THEN 1 ELSE 0 END AS is_primary,
				CASE WHEN c.column_default LIKE 'nextval%' THEN 1 ELSE 0 END AS is_auto_increment,
				CASE WHEN uni.column_name IS NOT NULL THEN 1 ELSE 0 END AS is_unique,
				c.column_default AS default,
				pg_catalog.col_description(
						(c.table_schema || '.' || c.table_name)::regclass,
						c.ordinal_position
				) AS comment
			FROM information_schema.columns c
				 LEFT JOIN (
							   SELECT
								   kcu.table_schema,
								   kcu.table_name,
								   kcu.column_name
							   FROM information_schema.table_constraints tc
									JOIN information_schema.key_column_usage kcu
									ON tc.constraint_schema = kcu.constraint_schema
										AND tc.constraint_name = kcu.constraint_name
							   WHERE tc.constraint_type = 'PRIMARY KEY'
						   ) pk
				 ON c.table_schema = pk.table_schema
					 AND c.table_name = pk.table_name
					 AND c.column_name = pk.column_name
				 LEFT JOIN (
							   SELECT
								   kcu.table_schema,
								   kcu.table_name,
								   kcu.column_name
							   FROM information_schema.table_constraints tc
									JOIN information_schema.key_column_usage kcu
									ON tc.constraint_schema = kcu.constraint_schema
										AND tc.constraint_name = kcu.constraint_name
							   WHERE tc.constraint_type = 'UNIQUE'
						   ) uni
				 ON c.table_schema = uni.table_schema
					 AND c.table_name = uni.table_name
					 AND c.column_name = uni.column_name
			WHERE c.table_catalog = ?
			ORDER BY c.table_name, c.ordinal_position;
		`,
		"oracle": `
			SELECT
				tc.COLUMN_ID AS "sort"
			  , NULL AS "database_name" -- Oracle 中通常不直接对应 TABLE_CATALOG
			  , tc.OWNER AS "schema_name"
			  , tc.TABLE_NAME AS "table_name"
			  , tc.COLUMN_NAME AS "column_name"
			  , tc.DATA_TYPE AS "data_type"
			  , tc.DATA_LENGTH AS "length"
			  , tc.DATA_PRECISION AS "precision"
			  , tc.DATA_SCALE AS "scale"
			  , CASE tc.NULLABLE WHEN 'Y' THEN 1 ELSE 0 END AS "nullable"
			  , CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END AS "is_primary"
			  , CASE WHEN idc.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END AS "is_auto_increment" -- 适用于 Oracle 12c 及以上版本
			  , CASE WHEN uc.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END AS "is_unique"
			  , tc.DATA_DEFAULT AS "default"
			  , cc.COMMENTS AS "comment"
			FROM ALL_TAB_COLUMNS tc
				 LEFT JOIN (
							   SELECT
								   ccu.TABLE_NAME
								 , ccu.COLUMN_NAME
							   FROM ALL_CONSTRAINTS cons
									JOIN ALL_CONS_COLUMNS ccu
									ON cons.CONSTRAINT_NAME = ccu.CONSTRAINT_NAME AND cons.OWNER = ccu.OWNER
							   WHERE
									 cons.CONSTRAINT_TYPE = 'P'
								 AND cons.OWNER = ?
						   ) pk
				 ON tc.TABLE_NAME = pk.TABLE_NAME AND tc.COLUMN_NAME = pk.COLUMN_NAME
				 LEFT JOIN (
							   SELECT
								   ccu.TABLE_NAME
								 , ccu.COLUMN_NAME
							   FROM ALL_CONSTRAINTS cons
									JOIN ALL_CONS_COLUMNS ccu
									ON cons.CONSTRAINT_NAME = ccu.CONSTRAINT_NAME AND cons.OWNER = ccu.OWNER
							   WHERE
									 cons.CONSTRAINT_TYPE = 'U'
								 AND cons.OWNER = ?
			--                      AND cons.TABLE_NAME = :table_name
						   ) uc
				 ON tc.TABLE_NAME = uc.TABLE_NAME AND tc.COLUMN_NAME = uc.COLUMN_NAME
				 LEFT JOIN ALL_TAB_IDENTITY_COLS idc
				 ON tc.TABLE_NAME = idc.TABLE_NAME AND tc.COLUMN_NAME = idc.COLUMN_NAME AND tc.OWNER = idc.OWNER
				 LEFT JOIN ALL_COL_COMMENTS cc
				 ON tc.TABLE_NAME = cc.TABLE_NAME AND tc.COLUMN_NAME = cc.COLUMN_NAME AND tc.OWNER = cc.OWNER
			WHERE
				tc.OWNER = UPPER(?)
			--   AND tc.TABLE_NAME = :table_name
			ORDER BY
				"database_name"
			  , "schema_name"
			  , "table_name"
			  , "sort"
		`,
		"sqlite": `
			SELECT 
				m.rowid AS 'sort',
				'main' AS 'database_name',
				'main' AS 'schema_name',
				m.name AS 'table_name',
				p.name AS 'column_name',
				p.type AS 'data_type',
				CASE 
					WHEN p.type LIKE '%CHAR%' OR p.type LIKE '%TEXT%' THEN 
						CAST(SUBSTR(p.type, INSTR(p.type, '(') + 1, INSTR(p.type, ')') - INSTR(p.type, '(') - 1) AS INTEGER)
					ELSE NULL 
				END AS 'length',
				CASE 
					WHEN p.type LIKE 'DECIMAL%' OR p.type LIKE 'NUMERIC%' THEN
						CAST(SUBSTR(p.type, INSTR(p.type, '(') + 1, INSTR(p.type, ',') - INSTR(p.type, '(') - 1) AS INTEGER)
					ELSE NULL 
				END AS 'precision',
				10 AS 'radix',
				CASE 
					WHEN p.type LIKE 'DECIMAL%' OR p.type LIKE 'NUMERIC%' THEN
						CAST(SUBSTR(p.type, INSTR(p.type, ',') + 1, INSTR(p.type, ')') - INSTR(p.type, ',') - 1) AS INTEGER)
					ELSE NULL 
				END AS 'scale',
				CASE WHEN p."notnull" = 0 THEN 1 ELSE 0 END AS 'nullable',
				CASE WHEN p.pk > 0 THEN 1 ELSE 0 END AS 'is_primary',
				CASE WHEN p.type LIKE '%INT%' AND p.pk > 0 THEN 1 ELSE 0 END AS 'is_auto_increment',
				CASE WHEN ui.column_name IS NOT NULL THEN 1 ELSE 0 END AS 'is_unique',
				p.dflt_value AS 'default',
				NULL AS 'comment'
			FROM sqlite_master m
			JOIN pragma_table_info(m.name) p ON 1=1
			LEFT JOIN (
				-- 获取唯一约束的字段信息
				SELECT 
					il.name AS table_name,
					ii.name AS column_name
				FROM sqlite_master m
				JOIN pragma_index_list(m.name) il ON 1=1
				JOIN pragma_index_info(il.name) ii ON 1=1
				WHERE m.type = 'table'
				AND il.origin = 'u'  -- 'u' 表示唯一约束[1](@ref)
			) ui ON m.name = ui.table_name AND p.name = ui.column_name
			WHERE m.type = 'table'
-- 			AND m.name = ?
			ORDER BY m.name, p.cid;
		`,
	}
)
