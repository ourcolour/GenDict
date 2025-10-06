package services

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"goDict/configs"
	"goDict/models"
	"goDict/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

var ExcelDatabaseRowMap = map[string]int{
	"DatabaseName": 3,
	"TableCount":   5,
	"TableMap":     8,
}

var ExcelTableRowMap = map[string]int{
	"TableName":  3,
	"TableType":  5,
	"Comment":    7,
	"ColumnList": 10,
}

// RENDERING_FUNC 渲染函数map
var RENDERING_FUNC = map[string]func(dbConfig *configs.DatabaseConfig, templateData interface{}, outputDirPath string, overwrite bool, total int, current int) (string, error){
	"md":   renderingMarkdown,
	"xlsx": renderingExcel,
}

// rendering 生成markdown
func (this *DbDictService) rendering(dbConfig *configs.DatabaseConfig, format string, templateData interface{}, outputDirPath string, overwrite bool, total int, current int) (string, error) {
	// 获取处理函数
	renderingFunc := RENDERING_FUNC[format]
	if nil == renderingFunc {
		return "", errors.New("不支持的格式")
	}

	// 调用函数渲染
	return renderingFunc(dbConfig, templateData, outputDirPath, overwrite, total, current)
}

// renderingMarkdown 渲染markdown
func renderingMarkdown(dbConfig *configs.DatabaseConfig, templateData interface{}, outputDirPath string, overwrite bool, total int, current int) (string, error) {
	/* 获取基本信息 */
	// 格式
	var format = "md"
	// 是否数据库
	var isDatabase = false
	// 数据名称
	var fileName string
	// 模板地址
	var templatePath string
	// 根据传入实体类型获取信息
	if _, ok := templateData.(*models.DatabaseInfo); ok {
		// 通过反射获取“DatabaseName”属性
		reflectValue, err := utils.ReflectFieldValue(templateData.(*models.DatabaseInfo), "DatabaseName")
		if nil != err {
			return "", err
		}
		fileName = reflectValue.String()
		// 指定模板地址
		templatePath = fmt.Sprintf("templates/db_dict_database.%s", format)
		// 标记为数据库
		isDatabase = true
	} else if _, ok := templateData.(*models.TableInfo); ok {
		// 通过反射获取“DatabaseName”属性
		reflectValue, err := utils.ReflectFieldValue(templateData.(*models.TableInfo), "DatabaseName")
		if nil != err {
			return "", err
		}
		fileName = reflectValue.String()
		// 指定模板地址
		templatePath = fmt.Sprintf("templates/db_dict_table.%s", format)
	} else {
		return "", fmt.Errorf("不支持的数据类型")
	}

	// SQLite 保存文件名取原始文件名
	if "SQLite" == dbConfig.Type {
		// 原始文件名
		srcFileName := filepath.Base(dbConfig.Host)
		// 去掉扩展名
		nameWithoutExt := strings.TrimSuffix(srcFileName, filepath.Ext(srcFileName))
		// 更新
		fileName = nameWithoutExt
	}

	/* 根据模板生成内容 */
	/*// 读取模板
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}*/
	// 读取模板
	t, err := template.ParseFS(templateFiles, templatePath)
	if err != nil {
		return "", err
	}

	// 使用 bytes.Buffer 捕获输出
	var buf bytes.Buffer
	err = t.Execute(&buf, templateData)
	if err != nil {
		return "", err
	}
	// 返回生成的文本
	content := buf.String()

	// 创建目录
	if _, err := mkDir(outputDirPath); err != nil {
		return "", err
	}
	// 保存路径信息
	fileExt := "md"
	savePath := path.Join(outputDirPath, fmt.Sprintf("%s.%s", fileName, fileExt))

	// 如果文件存在
	if utils.FileExists(savePath) {
		// 如果是数据库，并且不允许覆盖
		if isDatabase {
			// 不允许覆盖
			if !overwrite {
				return "", errors.New("文件已存在")
			}

			// 删除源文件
			if err := os.Remove(savePath); err != nil {
				return "", err
			}
		}
	}

	// 打开文件
	file, err := os.OpenFile(savePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if nil != err {
		return "", err
	}
	defer file.Close()

	// 写入文件
	_, err = file.WriteString(content)
	if err != nil {
		return "", err
	}

	return savePath, err
}

// 渲染excel
func renderingExcel(dbConfig *configs.DatabaseConfig, templateData interface{}, outputDirPath string, overwrite bool, total int, current int) (string, error) {
	/* 获取基本信息 */
	// 格式
	var format = "xlsx"
	// 是否数据库
	var isDatabase = false
	// 数据名称
	var fileName string
	// 模板地址
	var templatePath string
	// 根据传入实体类型获取信息
	if _, ok := templateData.(*models.DatabaseInfo); ok {
		// 标记为数据库
		isDatabase = true
		// 通过反射获取“DatabaseName”属性
		reflectValue, err := utils.ReflectFieldValue(templateData.(*models.DatabaseInfo), "DatabaseName")
		if nil != err {
			return "", err
		}
		fileName = reflectValue.String()
		// 指定模板地址（Excel使用同一个模板文件）
		templatePath = fmt.Sprintf("templates/db_dict_database.%s", format)
	} else if _, ok := templateData.(*models.TableInfo); ok {
		// 通过反射获取“DatabaseName”属性
		reflectValue, err := utils.ReflectFieldValue(templateData.(*models.TableInfo), "DatabaseName")
		if nil != err {
			return "", err
		}
		fileName = reflectValue.String()
		// 指定模板地址（Excel使用同一个模板文件）
		templatePath = fmt.Sprintf("templates/db_dict_database.%s", format)
	} else {
		return "", fmt.Errorf("不支持的数据类型")
	}

	/* 创建目录 */
	// 创建目录
	if _, err := mkDir(outputDirPath); err != nil {
		return "", err
	}
	// 保存路径信息
	fileExt := "xlsx"
	savePath := path.Join(outputDirPath, fmt.Sprintf("%s.%s", fileName, fileExt))

	/* 根据模板生成内容 */
	// 如果是数据库，并且不允许覆盖
	if isDatabase {
		// 如果文件存在
		if utils.FileExists(savePath) { // 不允许覆盖
			if !overwrite {
				return "", errors.New("文件已存在")
			}

			// 删除源文件
			if err := os.Remove(savePath); err != nil {
				return "", err
			}
		}

		// 复制模板文件到保存路径
		/*_, err := utils.FileCopy(templatePath, savePath)
		if err != nil {
			return "", err
		}*/
		templateFile, err := templateFiles.ReadFile(templatePath)
		if err != nil {
			return "", err
		}
		err = os.WriteFile(savePath, templateFile, 0644)
		if err != nil {
			return "", err
		}
	}

	// 打开文件
	doc, err := excelize.OpenFile(savePath)
	if nil != err {
		return "", err
	}

	/* 生成表格 / 数据库 */
	if isDatabase {
		err = renderingExcelDatabase(dbConfig, templateData, doc, total, current)
	} else {
		err = renderingExcelTable(dbConfig, templateData, doc, total, current)
	}
	if nil != err {
		return "", err
	}

	// 另存为
	err = doc.Save()
	if nil != err {
		return "", err
	}

	return savePath, nil
}

// renderingExcelDatabase 渲染Excel数据库
func renderingExcelDatabase(dbConfig *configs.DatabaseConfig, templateData interface{}, doc *excelize.File, total int, current int) error {
	// 转换值
	objInfo := templateData.(*models.DatabaseInfo)

	// 获取模板sheet索引
	tplSheetIndex, err := doc.GetSheetIndex("模板-database")
	// 创建当前sheet
	sheetName := filepath.Base("首页")
	newSheetIndex, err := doc.NewSheet(sheetName)
	// 复制模板并放到最后
	if err = doc.CopySheet(tplSheetIndex, newSheetIndex); err != nil {
		return err
	}
	// 激活新sheet
	doc.SetActiveSheet(newSheetIndex)

	/* 填写数据 */
	doc.SetCellValue(sheetName, fmt.Sprintf("B%d", ExcelDatabaseRowMap["DatabaseName"]), objInfo.DatabaseName)
	doc.SetCellValue(sheetName, fmt.Sprintf("B%d", ExcelDatabaseRowMap["TableCount"]), fmt.Sprintf("%d / %d", objInfo.GetSelectedTableCount(), objInfo.GetTableCount()))

	// 定义边框样式
	tableStyle1, err := doc.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center", // 水平居中
			Vertical:   "center", // 垂直居中
		},
	})

	// 处理列表
	tableRowNo := ExcelDatabaseRowMap["TableMap"]
	// 选中的表信息map
	selectedTableInfoMap := objInfo.GetSelectedTableMap()

	// 遍历处理每一个表格
	var idx int = 0
	for _, tblInfo := range selectedTableInfoMap {
		// 行号
		tableRow := tableRowNo + idx
		// 设置内容
		doc.SetCellValue(sheetName, fmt.Sprintf("B%d", tableRow), tblInfo.TableName)
		doc.SetCellHyperLink(sheetName, fmt.Sprintf("B%d", tableRow), tblInfo.TableName+"!A1", "Location")
		doc.SetCellValue(sheetName, fmt.Sprintf("C%d", tableRow), tblInfo.TableType)
		doc.SetCellValue(sheetName, fmt.Sprintf("D%d", tableRow), tblInfo.Comment)

		// 索引增加
		idx++
	}

	// 设置样式
	doc.SetCellStyle(sheetName, fmt.Sprintf("B%d", tableRowNo), fmt.Sprintf("D%d", tableRowNo+len(selectedTableInfoMap)-1), tableStyle1)

	return nil
}

// renderingExcelTable 渲染Excel表格
func renderingExcelTable(dbConfig *configs.DatabaseConfig, templateData interface{}, doc *excelize.File, total int, current int) error {
	// 转换值
	objInfo := templateData.(*models.TableInfo)

	// 获取模板sheet索引
	tplSheetIndex, err := doc.GetSheetIndex("模板-table")
	// 创建当前sheet
	sheetName := filepath.Base(objInfo.TableName)
	newSheetIndex, err := doc.NewSheet(sheetName)
	// 复制模板并放到最后
	if err = doc.CopySheet(tplSheetIndex, newSheetIndex); err != nil {
		return err
	}
	// 激活新sheet
	doc.SetActiveSheet(newSheetIndex)

	/* 填写数据 */
	// 返回链接
	doc.SetCellHyperLink(sheetName, "A1", "首页!A1", "Location")
	// 表头
	doc.SetCellValue(sheetName, "B1", "数据表："+objInfo.TableName)
	// 表名
	doc.SetCellValue(sheetName, fmt.Sprintf("B%d", ExcelTableRowMap["TableName"]), objInfo.TableName)
	// 类型
	doc.SetCellValue(sheetName, fmt.Sprintf("B%d", ExcelTableRowMap["TableType"]), objInfo.TableType)
	// 说明
	doc.SetCellValue(sheetName, fmt.Sprintf("B%d", ExcelTableRowMap["Comment"]), objInfo.Comment)

	// 定义边框样式
	tableStyle1, err := doc.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center", // 水平居中
			Vertical:   "center", // 垂直居中
		},
		//Fill: &excelize.Fill{
		//	Type:    "pattern",
		//	Color:   []string{"#D9D9D9"},
		//	Pattern: 1,
		//},
	})

	// 处理列表
	yesNoMap := map[bool]string{true: "✔", false: "-"}
	tableRowNo := ExcelTableRowMap["ColumnList"]
	columnList := objInfo.ColumnList

	// 新增对应该行数
	doc.InsertRows(sheetName, tableRowNo, len(columnList)-1)

	// 填写每一行数据
	for idx, colValue := range columnList {
		// 行号
		tableRow := tableRowNo + idx
		// 设置内容
		doc.SetCellValue(sheetName, fmt.Sprintf("B%d", tableRow), colValue.ColumnName)
		doc.SetCellValue(sheetName, fmt.Sprintf("C%d", tableRow), colValue.DataType)
		doc.SetCellValue(sheetName, fmt.Sprintf("D%d", tableRow), fmt.Sprintf("%d, %d, %d", colValue.Precision, colValue.Scale, colValue.Radix))
		doc.SetCellValue(sheetName, fmt.Sprintf("E%d", tableRow), yesNoMap[colValue.Nullable])
		doc.SetCellValue(sheetName, fmt.Sprintf("F%d", tableRow), colValue.Default)
		doc.SetCellValue(sheetName, fmt.Sprintf("G%d", tableRow), yesNoMap[colValue.IsPrimary])
		doc.SetCellValue(sheetName, fmt.Sprintf("H%d", tableRow), yesNoMap[colValue.IsAutoIncrement])
		doc.SetCellValue(sheetName, fmt.Sprintf("I%d", tableRow), yesNoMap[colValue.IsUnique])
		doc.SetCellValue(sheetName, fmt.Sprintf("J%d", tableRow), colValue.Comment)
	}
	// 设置文字居中
	doc.SetCellStyle(sheetName, fmt.Sprintf("B%d", tableRowNo), fmt.Sprintf("J%d", tableRowNo+len(columnList)-1), tableStyle1)

	// 往下移动数据列行数+2行
	tableRowNo += len(columnList) + 2

	// 索引列表
	indexList := objInfo.IndexList
	for idx, colValue := range indexList {
		// 行号
		tableRow := tableRowNo + idx
		// 设置内容
		doc.SetCellValue(sheetName, fmt.Sprintf("B%d", tableRow), colValue.IndexName)
		doc.SetCellValue(sheetName, fmt.Sprintf("C%d", tableRow), colValue.ColumnNames)
		doc.SetCellValue(sheetName, fmt.Sprintf("G%d", tableRow), colValue.IndexType)
		doc.SetCellValue(sheetName, fmt.Sprintf("H%d", tableRow), yesNoMap[colValue.IsPrimary])
		doc.SetCellValue(sheetName, fmt.Sprintf("I%d", tableRow), yesNoMap[colValue.IsUnique])
		doc.SetCellValue(sheetName, fmt.Sprintf("J%d", tableRow), colValue.IndexComment)

		// 合并“字段”单元格
		doc.MergeCell(sheetName, fmt.Sprintf("C%d", tableRow), fmt.Sprintf("F%d", tableRow))
	}
	// 设置文字居中
	doc.SetCellStyle(sheetName, fmt.Sprintf("B%d", tableRowNo), fmt.Sprintf("J%d", tableRowNo+len(indexList)-1), tableStyle1)

	// 当前为最后一个表时
	if current == total {
		// 移除模板页
		doc.DeleteSheet("模板-database")
		doc.DeleteSheet("模板-table")
		// 激活首页
		doc.SetActiveSheet(0)
	}

	return nil
}

// mkDir 创建目录
func mkDir(outputDirPath string) (string, error) {
	if _, err := os.Stat(outputDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDirPath, os.ModePerm); err != nil {
			return "", err
		}
	}
	return "", nil
}
