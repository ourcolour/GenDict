##### 名称：{{.TableName}}

##### 类型：{{if eq "table" .TableType}}表格 (table){{else}}视图 (view){{end}}

##### 说明：`{{if .Comment}}{{.Comment}}{{else}}（无）{{end}}`

| 字段名                   | 类型 | 长度, 精度 | 允许空值                                                                              | 默认值 | 主键 | 自增 | 唯一 | 说明 |
|-----------------------|------|-----------|-----------------------------------------------------------------------------------|--------|------|------|------|------|
{{range .ColumnList}} | {{.ColumnName}} | {{.DataType}} | {{if .Precision}}{{.Precision}}, {{.Radix}}, {{.Scale}}{{else}}{{.Length}}{{end}} | {{if .Nullable}}✓{{else}}-{{end}} | {{.Default}} | {{if .IsPrimary}}✓{{else}}-{{end}} | {{if .IsAutoIncrement}}✓{{else}}-{{end}} | {{if .IsUnique}}✓{{else}}-{{end}} | {{if .Comment}}{{.Comment}}{{else}}-{{end}} |
{{end}}

| 索引                   | 字段 | 唯一 | 主键 | 类型 | 说明 |
|----------------------|------|------|------|------|------|
{{range .IndexList}} | {{.IndexName}}  | {{.ColumnNames}} | {{if .IsUnique}}✓{{else}}-{{end}} | {{if .IsPrimary}}✓{{else}}-{{end}} | {{.IndexType}} | {{.IndexComment}} |
{{end}}

----------
