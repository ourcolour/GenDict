# 数据库字典

### 库名：{{.DatabaseName}}

### 数量：{{.TableCount}}

### 清单：

| 表名                                        | 类型 | 说明                                                                   |
|-------------------------------------------|----|----------------------------------------------------------------------|
 {{range $tableName, $table := .TableMap}} | [{{$table.TableName}}](#名称：{{$table.TableName}}) | {{if eq "table" $table.TableType}}表格 (table){{else}}视图 (view){{end}} | {{if $table.Comment}}{{$table.Comment}}{{else}}-{{end}} |
{{end}}

----------

