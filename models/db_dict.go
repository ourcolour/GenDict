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
	DatabaseName string `json:"database_name"`
	TableName    string `json:"table_name"`
	TableType    string `json:"table_type"`
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
	TableMap      map[string]TableInfo
	TableNameList []string
	// 选中的表名
	selectedTableNameList []string
}

// NewDatabaseInfo 创建数据库信息结构体
func NewDatabaseInfo(dbName string, tblMap map[string]TableInfo, selectedTableNameList []string) *DatabaseInfo {
	result := &DatabaseInfo{}
	result.DatabaseName = dbName
	result.TableMap = tblMap

	// 从tableMap中提取表名
	result.TableNameList = make([]string, 0)
	for tblName, _ := range tblMap {
		result.TableNameList = append(result.TableNameList, tblName)
	}

	// 如果没有指定，返回全部
	if nil == result.selectedTableNameList {
		result.selectedTableNameList = result.TableNameList
		return result
	}

	// 选中的表名
	result.selectedTableNameList = make([]string, 0)
	// 确保选中的表名有效
	for _, tblName := range selectedTableNameList {
		if _, ok := tblMap[tblName]; ok {
			result.selectedTableNameList = append(result.selectedTableNameList, tblName)
		}
	}

	return result
}

// GetTableCount 获取表数量
func (this *DatabaseInfo) GetTableCount() int {
	return len(this.TableNameList)
}

// GetSelectedTableNameList 获取选中的表名列表
func (this *DatabaseInfo) GetSelectedTableCount() int {
	return len(this.selectedTableNameList)
}

// GetSelectedTableNameList 获取选中的表名列表
func (this *DatabaseInfo) GetSelectedTableNameList() []string {
	// Args
	if nil == this.selectedTableNameList {
		this.selectedTableNameList = make([]string, 0)
	}

	return this.selectedTableNameList
}

// SetSelectedTableNameList 设置选中的表名列表
func (this *DatabaseInfo) SetSelectedTableNameList(selectedTableNameList []string) {
	this.selectedTableNameList = selectedTableNameList
	// 防止空指针
	if nil == this.selectedTableNameList {
		this.selectedTableNameList = make([]string, 0)
	}
}

// GetSelectedTableMap 获取选中的表信息map
func (this *DatabaseInfo) GetSelectedTableMap() map[string]*TableInfo {
	result := make(map[string]*TableInfo)

	// 选中的表名列表
	selectedTableNameList := this.GetSelectedTableNameList()
	if nil == selectedTableNameList || 1 > len(selectedTableNameList) {
		return result
	}

	// 选中的表名列表
	for _, tblName := range selectedTableNameList {
		// 当前表信息
		tableInfo, ok := this.TableMap[tblName]
		if !ok {
			continue
		}

		// 添加表信息
		result[tblName] = &tableInfo
	}

	return result
}
