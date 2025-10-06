package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log/slog"
	"strings"
	"sync"
	"time"
)

var (
	// 类型值列表
	TypeValueList []string = []string{
		"all",
		"table",
		"view",
	}
	// 类型标签列表
	TypeLabelList []string = []string{
		"全部 (All)",
		"表 (table)",
		"视图 (view)",
	}
	/*	// 类型map
		TypeMap map[string]string = map[string]string{
			TypeLabelList[0]: TypeValueList[0],
			TypeLabelList[1]: TypeValueList[1],
			TypeLabelList[2]: TypeValueList[2],
		}*/

	// 防抖函数延迟（毫秒）
	DebounceDelay time.Duration = 300 * time.Millisecond
)

type SearchView struct {
	*BaseView

	app    *fyne.App
	window *fyne.Window

	// 防抖函数
	debounceSearch func(...interface{})
	// 完成事件
	OnFinished func(selectedTableNameList []string)

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
	// 过滤数据
	tblFilteredData [][]string
	// 原始数据
	tblRawData [][]string
	// 表选中状态map(key:表名, value:选中状态)
	tblSelectedMap map[string]bool

	// 复选框map
	checkBoxMap map[string]*widget.Check
}

// NewSearchView 创建搜索视图
func NewSearchView(app *fyne.App, tableInfoMap map[string]string) *SearchView {
	view := &SearchView{}

	// 初始化窗口
	view.init(app, tableInfoMap)

	return view
}

// NewSearchViewWithOnFinishedEvent 创建搜索视图（包含回调函数）
func NewSearchViewWithOnFinishedEvent(app *fyne.App, tableInfoMap map[string]string, fn func([]string)) *SearchView {
	view := NewSearchView(app, tableInfoMap)
	view.OnFinished = fn
	return view
}

// ------------------------------
// 初始化
// ------------------------------
// init 初始化
func (this *SearchView) init(app *fyne.App, tableInfoMap map[string]string) {
	// 初始化防抖函数
	this.debounceSearch = this.debounce(DebounceDelay, func(args ...interface{}) {
		if len(args) >= 2 {
			dbTypeIndex := args[0].(int)
			keyword := args[1].(string)
			this.performSearch(dbTypeIndex, keyword)
		}
	})

	// 类型标签列表
	TypeLabelList = []string{
		I("search-view.ui.selType.optionList.option0"),
		I("search-view.ui.selType.optionList.option1"),
		I("search-view.ui.selType.optionList.option2"),
	}

	// 应用
	this.app = app
	// 窗口
	window := (*this.app).NewWindow(I("search-view.ui.window.title"))
	this.window = &window
	(*this.window).Resize(fyne.NewSize(580, 480))
	// 初始化数据
	this.initData(tableInfoMap)
	// 设置内容
	(*this.window).SetContent(this.initUI())

	// 选中第一个
	this.selType.SetSelectedIndex(0)
	// 更新选中数量
	this.updateCount()
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
		this.tblRawData = append(this.tblRawData, []string{"0", tblName, tblType})
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
	this.selType = widget.NewSelect(TypeLabelList, this.selType_onChange)
	// 关键词
	this.txtKeyword = widget.NewEntry()
	this.txtKeyword.OnChanged = this.txtKeyword_onChanged
	// 选中按钮
	this.btnAdd = widget.NewButtonWithIcon(I("search-view.ui.btnAdd.text"), theme.ContentAddIcon(), this.btnAdd_onClick)
	// 移除按钮
	this.btnRemove = widget.NewButtonWithIcon(I("search-view.ui.btnRemove.text"), theme.ContentRemoveIcon(), this.btnRemove_onClick)

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
		widget.NewLabel(I("search-view.ui.top.right.label.text")),
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
	this.lblSelected = widget.NewLabel(Id("search-view.ui.lblSelected.text", map[string]interface{}{"Count": 0}))
	// 总数量
	this.lblTotal = widget.NewLabel(Id("search-view.ui.lblTotal.text", map[string]interface{}{"Count": 0}))
	// 完成按钮
	this.btnFinish = widget.NewButtonWithIcon(I("search-view.ui.btnFinish.text"), theme.MediaPlayIcon(), this.btnFinish_onClick)
	this.btnFinish.Importance = widget.HighImportance

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
func (this *SearchView) selType_onChange(label string) {
	slog.Debug("selType_onChange", "value", label)

	// 使用防抖函数而不是直接调用搜索
	this.debounceSearch(this.selType.SelectedIndex(), this.txtKeyword.Text)
}

// txtKeyword_onChanged 输入框内容改变事件
func (this *SearchView) txtKeyword_onChanged(value string) {
	slog.Debug("txtKeyword_onChanged", "value", value)

	// 使用防抖函数而不是直接调用搜索
	this.debounceSearch(this.selType.SelectedIndex(), this.txtKeyword.Text)
}

// btnAdd_onClick 选中按钮点击事件
func (this *SearchView) btnAdd_onClick() {
	// 遍历表格原始数据
	for _, row := range this.tblFilteredData {
		// 名称
		curName := row[1]

		this.tblSelectedMap[curName] = true
	}

	// 刷新
	this.tblResult.Refresh()
	// 更新选中数量、总数量
	this.updateCount()
}

// btnRemove_onClick 移除按钮点击事件
func (this *SearchView) btnRemove_onClick() {
	// 遍历表格原始数据
	for _, row := range this.tblFilteredData {
		// 名称
		curName := row[1]

		this.tblSelectedMap[curName] = false
	}

	// 刷新
	this.tblResult.Refresh()
	// 更新选中数量、总数量
	this.updateCount()
}

// tblResult_length 表格单元格长度
func (this *SearchView) tblResult_length() (rows int, cols int) {
	// 列数
	cols = 3

	// 没有数据
	if nil == this.tblFilteredData {
		rows = 0
	} else {
		// 行数 = 数据行数 + 表头行(1行)
		rows = len(this.tblFilteredData) + 1
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
	if 0 == tableCellId.Row {
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
		control, ok := ctn.Objects[0].(*widget.Label)
		if !ok {
			// 移除内容
			ctn.RemoveAll()
			// 添加控件
			control = widget.NewLabel(I("search-view.ui.tblResult.headerCell.lblChoose.text"))
			// 添加到容器
			ctn.Add(control)
		} else {
			// 设置文本
			control.SetText(I("search-view.ui.tblResult.headerCell.lblChoose.text"))
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
			control = widget.NewLabel(I("search-view.ui.tblResult.headerCell.lblType.text"))
			// 添加到容器
			ctn.Add(control)
		} else {
			// 设置文本
			control.SetText(I("search-view.ui.tblResult.headerCell.lblType.text"))
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
			control = widget.NewLabel(I("search-view.ui.tblResult.headerCell.lblName.text"))
			// 添加到容器
			ctn.Add(control)
		} else {
			// 设置文本
			control.SetText(I("search-view.ui.tblResult.headerCell.lblName.text"))
		}
		break
	}

	// 刷新
	ctn.Refresh()
}

// 更新数据单元格
func (this *SearchView) tblResult_updateDataCell(tableCellId widget.TableCellID, ctn *fyne.Container) {
	if this.tblFilteredData == nil || tableCellId.Row-1 >= len(this.tblFilteredData) {
		return
	}

	// 当前数据
	data := this.tblFilteredData[tableCellId.Row-1]
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

	// 单选
	this.tblSelectedMap[tableName] = checked
	// 更新选中数量、总数量
	this.updateCount()
}

// btnFinish_onClick 完成按钮点击事件
func (this *SearchView) btnFinish_onClick() {
	// 触发回调
	this.triggerOnFinishedEvent()
}

// ------------------------------
// Functions
// ------------------------------
// Show 显示
func (this *SearchView) Show() {
	// 显示窗口
	(*this.window).Show()
	// 窗口置顶
	(*this.window).RequestFocus()
}

// close 关闭窗口
func (this *SearchView) close() {
	// 关闭窗口
	(*this.window).Close()
}

// search 搜索
func (this *SearchView) searchData(objType string, keyword string) [][]string {
	// 清空过滤数据
	result := [][]string{}

	// 遍历表格原始数据
	for _, row := range this.tblRawData {
		// 索引
		//curIdx := row[0]
		// 名称
		curName := row[1]
		// 类型
		curType := row[2]

		// 是否匹配
		isMatched := true

		// 不是所有
		if "all" != objType {
			// 匹配类型
			isMatched = isMatched && objType == curType
		}

		// 忽略空字符串
		if "" != strings.TrimSpace(keyword) {
			// 匹配名称
			isMatched = isMatched && strings.Contains(curName, keyword)
		}

		// 如果匹配
		if isMatched {
			// 将结果添加到过滤数据中
			result = append(result, row)
		}
	}

	// 返回过滤数据
	return result
}

// performSearch 执行搜索并更新UI
func (this *SearchView) performSearch(objTypeIndex int, keyword string) {
	// 选中索引转换为选中值
	objType := TypeValueList[objTypeIndex]
	// 搜索数据
	this.tblFilteredData = this.searchData(objType, keyword)
	// 刷新表格
	fyne.DoAndWait(func() {
		this.tblResult.Refresh()
	})
}

// debounce 防抖函数函数
func (sv *SearchView) debounce(interval time.Duration, fn func(...interface{})) func(...interface{}) {
	var timer *time.Timer
	var lastArgs []interface{}
	// 用于保证并发安全
	var mu sync.Mutex

	return func(args ...interface{}) {
		mu.Lock()
		defer mu.Unlock()

		// 保存参数
		lastArgs = args

		// 如果已有定时器在运行，则停止它
		if timer != nil {
			timer.Stop()
		}

		// 设置新的定时器，到期后使用最后一次调用保存的参数执行目标函数
		timer = time.AfterFunc(interval, func() {
			mu.Lock()
			defer mu.Unlock()
			fn(lastArgs...)
		})
	}
}

// updateCount 更新选中数量、总数量
func (this *SearchView) updateCount() {
	// 选中数量
	this.lblSelected.SetText(Id("search-view.ui.lblSelected.text", map[string]interface{}{"Count": this.getSelectedCount()}))
	// 总数量
	this.lblTotal.SetText(Id("search-view.ui.lblTotal.text", map[string]interface{}{"Count": len(this.tblNameList)}))
}

// getSelectedTableNameList 获取选中的表名列表
func (this *SearchView) getSelectedTableNameList() []string {
	result := []string{}

	// 找出选中值
	for tblName, selected := range this.tblSelectedMap {
		if selected {
			result = append(result, tblName)
		}
	}

	return result
}

// 获取选中的表数量
func (this *SearchView) getSelectedCount() int {
	return len(this.getSelectedTableNameList())
}

// triggerOnFinishedEvent 触发完成事件
func (this *SearchView) triggerOnFinishedEvent() {
	// 如果指定完成事件
	if nil == this.OnFinished {
		return
	}

	// 选中表名列表
	selectedTableNameList := this.getSelectedTableNameList()
	if nil == selectedTableNameList || 1 > len(selectedTableNameList) {
		dialog.ShowInformation(I("search-view.ui.dialog.info.title"), I("search-view.ui.dialog.msg.noSelected"), (*this.window))
		return
	}

	// 调用完成事件，返回选中状态map
	this.OnFinished(selectedTableNameList)

	// 关闭窗口
	this.close()
}
