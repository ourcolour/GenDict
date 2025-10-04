package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log/slog"
)

var (
	// 类型map
	TypeNameMap map[string]string = map[string]string{
		"all":   "全部 (All)",
		"table": "表 (table)",
		"view":  "视图 (view)",
	}
)

type SearchView struct {
	app    *fyne.App
	window *fyne.Window

	/* top */
	// 类型选择
	selType *widget.Select
	// 关键词
	txtKeyword *widget.Entry
	// 选中按钮
	btnAdd *widget.Button
	// 移除按钮
	btnRemove *widget.Button
	/* middle */
	// 搜索表格
	tblResult *widget.Table
	/* bottom */
	// 选中数量
	lblSelected *widget.Label
	// 总数量
	lblTotal *widget.Label
	// 完成按钮
	btnFinish *widget.Button

	// 表名列表
	tblNameList []string
	// 表选中状态map(key:表名, value:选中状态)
	tblSelectedMap map[string]bool
	// 表格数据
	tblData [][]string

	// 复选框map
	checkBoxMap map[string]*widget.Check
}

// NewSearchView 创建搜索视图
func NewSearchView(app *fyne.App, tableInfoMap map[string]string) *SearchView {
	view := &SearchView{}
	view.init(app, tableInfoMap)
	return view
}

// ------------------------------
// 初始化
// ------------------------------
// init 初始化
func (this *SearchView) init(app *fyne.App, tableInfoMap map[string]string) {
	// 应用
	this.app = app
	// 窗口
	window := (*this.app).NewWindow("搜索")
	this.window = &window
	(*this.window).Resize(fyne.NewSize(550, 480))
	// 初始化数据
	this.initData(tableInfoMap)
	// 设置内容
	(*this.window).SetContent(this.initUI())
}

// initData 初始化数据
func (this *SearchView) initData(tableInfoMap map[string]string) {
	// 表名列表
	this.tblNameList = []string{}
	// 表选中状态map
	this.tblSelectedMap = map[string]bool{}
	// 复选框map
	this.checkBoxMap = map[string]*widget.Check{}

	// 遍历
	for tblName, tblType := range tableInfoMap {
		// 提取名称列表
		this.tblNameList = append(this.tblNameList, tblName)
		// 选中状态
		this.tblSelectedMap[tblName] = false
		// 表格数据
		this.tblData = append(this.tblData, []string{"0", tblName, tblType})
	}

	return
}

// initUI 初始化UI
func (this *SearchView) initUI() *fyne.Container {
	// top
	topContainer := this.initUITop()
	// middle
	middleContainer := this.initUIMiddle()
	// bottom
	bottomContainer := this.initUIBottom()

	return container.NewPadded(
		container.NewBorder(
			topContainer,
			bottomContainer,
			nil,
			nil,
			middleContainer,
		),
	)
}

// initUITop 初始化顶部UI
func (this *SearchView) initUITop() *fyne.Container {
	// 类型下拉框
	this.selType = widget.NewSelect(this.getTypeValueList(), this.selType_onChange)
	this.selType.SetSelectedIndex(0)
	// 关键词
	this.txtKeyword = widget.NewEntry()
	this.txtKeyword.OnChanged = this.txtKeyword_onChanged
	// 选中按钮
	this.btnAdd = widget.NewButtonWithIcon("选中", theme.ContentAddIcon(), this.btnAdd_onClick)
	// 移除按钮
	this.btnRemove = widget.NewButtonWithIcon("移除", theme.ContentRemoveIcon(), this.btnRemove_onClick)

	// 左侧
	leftContrainer := container.NewBorder(
		nil,
		nil,
		this.selType,
		nil,
		this.txtKeyword,
	)
	// 右侧
	rightContainer := container.NewHBox(
		widget.NewLabel("将结果："),
		this.btnAdd,
		this.btnRemove,
	)

	return container.NewBorder(
		nil,
		nil,
		nil,
		rightContainer,
		leftContrainer,
	)
}

// initUIMiddle 初始化中部UI
func (this *SearchView) initUIMiddle() *fyne.Container {
	// 表格
	this.tblResult = widget.NewTable( // WithHeaders
		this.tblResult_length,
		this.tblResult_create,
		this.tblResult_update,
	)

	// 设置列狂
	this.tblResult.SetColumnWidth(0, 60)
	this.tblResult.SetColumnWidth(1, 200)
	this.tblResult.SetColumnWidth(2, 260)
	this.tblResult.Refresh()

	return container.NewStack(this.tblResult)
}

// initUIBottom 初始化底部UI
func (this *SearchView) initUIBottom() *fyne.Container {
	// 选中数量
	this.lblSelected = widget.NewLabel(fmt.Sprintf("选中: %d", len(this.tblSelectedMap)))
	// 总数量
	this.lblTotal = widget.NewLabel(fmt.Sprintf("总数: %d", len(this.tblNameList)))
	// 完成按钮
	this.btnFinish = widget.NewButtonWithIcon("完成", theme.ConfirmIcon(), this.btnFinish_onClick)

	// 左侧
	leftContainer := container.NewHBox(
		this.lblSelected,
		this.lblTotal,
	)

	return container.NewBorder(
		nil,
		nil,
		nil,
		this.btnFinish,
		leftContainer,
	)
}

// ------------------------------
// 事件处理
// ------------------------------
// selType_onChange 类型选择事件
func (this *SearchView) selType_onChange(value string) {
	slog.Debug("selType_onChange", "value", value)
}

// txtKeyword_onChanged 输入框内容改变事件
func (this *SearchView) txtKeyword_onChanged(value string) {
	slog.Debug("txtKeyword_onChanged", "value", value)
}

// btnAdd_onClick 选中按钮点击事件
func (this *SearchView) btnAdd_onClick() {
}

// btnRemove_onClick 移除按钮点击事件
func (this *SearchView) btnRemove_onClick() {
}

// tblResult_length 表格单元格长度
func (this *SearchView) tblResult_length() (rows int, cols int) {
	// 列数
	cols = 3

	// 没有数据
	if nil == this.tblData {
		rows = 0
	} else {
		// 行数 = 数据行数 + 表头行(1行)
		rows = len(this.tblData) + 1
	}

	return rows, cols
}

// tblResult_create 表格单元格创建
func (this *SearchView) tblResult_create() fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel(""))
}

// tblResult_update 表格单元格更新
func (this *SearchView) tblResult_update(tableCellId widget.TableCellID, obj fyne.CanvasObject) {
	slog.Debug("tblResult_update", "tableCellId", tableCellId)

	ctn, ok := obj.(*fyne.Container)
	if !ok {
		return
	}

	// 表头行处理
	if tableCellId.Row == 0 {
		this.tblResult_updateHeaderCell(tableCellId, ctn)
		return
	}

	// 数据行处理
	this.tblResult_updateDataCell(tableCellId, ctn)

	// 设置行高
	this.tblResult.SetRowHeight(tableCellId.Row, 30)
}

// 更新表头单元格
func (this *SearchView) tblResult_updateHeaderCell(tableCellId widget.TableCellID, ctn *fyne.Container) {
	if len(ctn.Objects) == 0 {
		return
	}

	// 不同列不同操作
	switch tableCellId.Col {
	// 复选框列
	case 0:
		// 控件
		control, ok := ctn.Objects[0].(*widget.Check)
		if !ok {
			// 移除内容
			ctn.RemoveAll()
			// 添加控件
			control = widget.NewCheck("", func(checked bool) {
				this.tblResult_onChecked(checked, "")
			})
			// 添加到容器
			ctn.Add(control)
		} else {
			// 设置状态
			control.SetChecked(false)
			// 绑定事件
			control.OnChanged = func(checked bool) {
				this.tblResult_onChecked(checked, "")
			}
		}
		break
	// 类型列
	case 1:
		// 控件
		control, ok := ctn.Objects[0].(*widget.Label)
		if !ok {
			// 移除内容
			ctn.RemoveAll()
			// 添加控件
			control = widget.NewLabel("类型")
			// 添加到容器
			ctn.Add(control)
		} else {
			// 设置文本
			control.SetText("类型")
		}
		break
	// 名称列
	case 2:
		// 控件
		control, ok := ctn.Objects[0].(*widget.Label)
		if !ok {
			// 移除内容
			ctn.RemoveAll()
			// 添加控件
			control = widget.NewLabel("名称")
			// 添加到容器
			ctn.Add(control)
		} else {
			// 设置文本
			control.SetText("名称")
		}
		break
	}

	// 刷新
	ctn.Refresh()
}

// 更新数据单元格
func (this *SearchView) tblResult_updateDataCell(tableCellId widget.TableCellID, ctn *fyne.Container) {
	if this.tblData == nil || tableCellId.Row-1 >= len(this.tblData) {
		return
	}

	// 当前数据
	data := this.tblData[tableCellId.Row-1]
	// 表名
	tableName := data[1]
	// 类型
	tableType := data[2]
	// 选中状态
	tableChecked := this.tblSelectedMap[tableName]

	switch tableCellId.Col {
	// 复选框列
	case 0:
		// 控件
		control, ok := ctn.Objects[0].(*widget.Check)
		if !ok {
			// 移除内容
			ctn.RemoveAll()
			// 添加控件
			control = widget.NewCheck("", func(checked bool) {
				this.tblResult_onChecked(checked, tableName)
			})
			// 添加到容器
			ctn.Add(control)
		} else {
			// 绑定事件
			control.OnChanged = func(checked bool) {
				this.tblResult_onChecked(checked, tableName)
			}
		}

		// 选中状态
		control.SetChecked(tableChecked)
		// 添加到map
		this.checkBoxMap[tableName] = control

		break
	// 名称列
	case 1:
		// 控件
		control, ok := ctn.Objects[0].(*widget.Label)
		if !ok {
			// 移除内容
			ctn.RemoveAll()
			// 添加控件
			control = widget.NewLabel(tableName)
			// 添加到容器
			ctn.Add(control)
		} else {
			// 添加控件
			control.SetText(tableName)
		}
		break
	// 类型列
	case 2:
		// 控件
		control, ok := ctn.Objects[0].(*widget.Label)
		if !ok {
			// 移除内容
			ctn.RemoveAll()
			// 添加控件
			control = widget.NewLabel(tableType)
			// 添加到容器
			ctn.Add(control)
		} else {
			// 添加控件
			control.SetText(tableType)
		}
		break
	}
	ctn.Refresh()
}

// tblResult_onChecked 表格复选框全选事件
func (this *SearchView) tblResult_onChecked(checked bool, tableName string) {
	slog.Info("tblResult_onChecked", "checked", checked, "tableName", tableName)

	// 全选
	if "" == tableName {
		for k, _ := range this.tblSelectedMap {
			// 更新数据
			this.tblSelectedMap[k] = checked
			// 复选框
			checkBox := this.checkBoxMap[k]
			if nil == checkBox {
				continue
			}
			// 移除事件
			checkBox.OnChanged = nil
			// 更新状态
			checkBox.SetChecked(checked)
			// 添加事件
			checkBox.OnChanged = func(checked bool) {
				this.tblResult_onChecked(checked, k)
			}
		}

		// 刷新
		//this.tblResult.Refresh()
		return
	}

	// 单选
	this.tblSelectedMap[tableName] = checked
}

// btnFinish_onClick 完成按钮点击事件
func (this *SearchView) btnFinish_onClick() {
}

// ------------------------------
// Functions
// ------------------------------
// Show 显示
func (this *SearchView) Show() {
	(*this.window).Show()
}

// getTypeValueList 获取类型列表
func (this *SearchView) getTypeValueList() []string {
	// 提取类型
	typeList := []string{}
	for _, v := range TypeNameMap {
		typeList = append(typeList, v)
	}
	return typeList
}
