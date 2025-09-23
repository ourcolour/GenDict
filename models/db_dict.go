package models

import "strings"

// IndexInfo 索引信息结构
type IndexInfo struct {
	DatabaseName string `json:"database_name"`
	SchemaName   string `json:"schema_name"`
	TableName    string `json:"table_name"`
	IndexName    string `json:"index_name"`
	ColumnNames  string `json:"column_name_list"`
	IsUnique     bool   `json:"is_unique"`
	IsPrimary    bool   `json:"is_primary"`
	IndexType    string `json:"index_type"`
	IndexComment string `json:"index_comment"`
}

func (this *IndexInfo) GetColumnNameList() []string {
	return strings.Split(this.ColumnNames, ", ")
}

// 表结构信息结构体
type ColumnInfo struct {
	Sort            int
	DatabaseName    string
	SchemaName      string
	TableName       string
	ColumnName      string
	DataType        string
	Length          int64
	Precision       int64
	Radix           int64
	Scale           int64
	Nullable        bool
	IsPrimary       bool
	IsAutoIncrement bool
	IsUnique        bool
	Default         string
	Comment         string
}

type DecimalSizeInfo struct {
	Precision int64
	Scale     int64
}

// 表对象类型
type TableType struct {
	TableName string `json:"table_name"`
	TableType string `json:"table_type"`
}

// 表信息结构体
type TableInfo struct {
	DatabaseName string
	TableName    string
	ColumnList   []*ColumnInfo
	Comment      string
	TableType    string // TABLE or VIEW
	IndexList    []*IndexInfo
}

// 数据库信息结构体
type DatabaseInfo struct {
	DatabaseName  string
	TableCount    int
	TableNameList []string
	TableMap      map[string]TableInfo
}
