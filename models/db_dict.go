package models

// 表结构信息结构体
type ColumnInfo struct {
	Name            string
	Type            string
	Nullable        bool
	Default         string
	Comment         string
	IsPrimary       bool
	IsAutoIncrement bool
	IsUnique        bool
	Length          int64
	DecimalSize     DecimalSizeInfo
	ColumnType      string
}

type DecimalSizeInfo struct {
	Precision int64
	Scale     int64
}

// 表信息结构体
type TableInfo struct {
	DatabaseName string
	TableName    string
	ColumnList   []ColumnInfo
	Comment      string
	TableType    string // TABLE or VIEW
}

// 数据库信息结构体
type DatabaseInfo struct {
	DatabaseName  string
	TableCount    int
	TableNameList []string
	TableMap      map[string]TableInfo
}
