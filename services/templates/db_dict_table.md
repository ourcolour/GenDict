##### 名称：{{.TableName}}

##### 类型：{{if eq "table" .TableType}}表格 (table){{else}}视图 (view){{end}}

##### 说明：`{{if .Comment}}{{.Comment}}{{else}}（无）{{end}}`

| 字段名 | 类型 | 长度, 精度 | 允许空值 | 默认值 | 主键 | 自增 | 唯一 | 说明 |
|--------|------|-----------|----------|--------|------|------|------|------|
{{range .ColumnList}}| {{.Name}} | {{.Type}} | {{if .DecimalSize.Precision}}{{.DecimalSize.Precision}}, {{.DecimalSize.Scale}}{{else}}{{.Length}}{{end}} | {{if .Nullable}}✓{{else}}-{{end}} | {{.Default}} | {{if .IsPrimary}}✓{{else}}-{{end}} | {{if .IsAutoIncrement}}✓{{else}}-{{end}} | {{if .IsUnique}}✓{{else}}-{{end}} | {{.Comment}} |
{{end}}

----------
