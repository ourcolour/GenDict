# 数据库字典/Database Dictionary

### 库名/Database：{{.DatabaseName}}

### 数量/Quantity：{{.TableCount}}

### 清单/List：

| 表名/Table                                  | 类型/Type                                          | 说明/Memo                                                              |
|-------------------------------------------|--------------------------------------------------|----------------------------------------------------------------------|
 {{range $tableName, $table := .TableMap}} | [{{$table.TableName}}](#名称：{{$table.TableName}}) | {{if eq "table" $table.TableType}}表格 (table){{else}}视图 (view){{end}} | {{if $table.Comment}}{{$table.Comment}}{{else}}-{{end}} |
{{end}}

----------

