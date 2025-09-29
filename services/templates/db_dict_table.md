##### 名称/Table：{{.TableName}}

##### 类型/Type：{{if eq "table" .TableType}}表格 (table){{else}}视图 (view){{end}}

##### 说明/Memo：`{{if .Comment}}{{.Comment}}{{else}}（无/Empty）{{end}}`

| 字段名/Field             | 类型/Type         | 长度, 精度/Len, Prec | 允许空/Nullable                                                                      | 默认值/Default                       | 主键/Primary   | 自增/AutoIncre                       | 唯一/Unique                                | 说明/Memo                           |
|-----------------------|-----------------|------------------|-----------------------------------------------------------------------------------|-----------------------------------|--------------|------------------------------------|------------------------------------------|-----------------------------------|
 {{range .ColumnList}} | {{.ColumnName}} | {{.DataType}}    | {{if .Precision}}{{.Precision}}, {{.Radix}}, {{.Scale}}{{else}}{{.Length}}{{end}} | {{if .Nullable}}✓{{else}}-{{end}} | {{.Default}} | {{if .IsPrimary}}✓{{else}}-{{end}} | {{if .IsAutoIncrement}}✓{{else}}-{{end}} | {{if .IsUnique}}✓{{else}}-{{end}} | {{if .Comment}}{{.Comment}}{{else}}-{{end}} |
{{end}}

| 索引/Index             | 字段/Field       | 唯一/Unique        | 主键/Primary                        | 类型/Type                            | 说明/Memo        |
|----------------------|----------------|------------------|-----------------------------------|------------------------------------|----------------|
 {{range .IndexList}} | {{.IndexName}} | {{.ColumnNames}} | {{if .IsUnique}}✓{{else}}-{{end}} | {{if .IsPrimary}}✓{{else}}-{{end}} | {{.IndexType}} | {{.IndexComment}} |
{{end}}

----------
