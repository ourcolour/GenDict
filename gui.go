package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"goDict/configs"
	"goDict/utils"
	"strconv"
	"strings"
)

var (
	// 数据库类型列表
	DbTypeList = []string{"MySQL", "SQLServer", "Oracle", "SQLite", "PostgresSQL"}
	// 数据库端口
	DbPortMap = map[string]int{
		"MySQL":     3306,
		"SQLServer": 1433,
		"Oracle":    5432,
		"SQLite":    0,
	}
	// 支付编码列表
	CharsetList = []string{"utf8mb4", "utf8", "gbk", "gb2312", "latin1"}
	// 输出格式列表
	OutputFormatList = []string{"xlsx", "md"}

	// 显示模式
	DisplayModeMap = map[string]map[string]bool{
		"MySQL":     DisplayModeNormal,
		"SQLServer": DisplayModeSQLServer,
		"Oracle":    DisplayModeNormal,
		"SQLite":    DisplayModeSQLite,
	}
	DisplayModeNormal = map[string]bool{
		"SelDbType":   true,
		"TxtHost":     true,
		"TxtPort":     true,
		"TxtUsername": true,
		"TxtPassword": true,
		"TxtDbName":   true,
		"SelCharset":  true,
	}
	DisplayModeSQLite = map[string]bool{
		"SelDbType":   true,
		"TxtHost":     true,
		"TxtPort":     false,
		"TxtUsername": false,
		"TxtPassword": false,
		"TxtDbName":   false,
		"SelCharset":  false,
	}
	DisplayModeSQLServer = map[string]bool{
		"SelDbType":   true,
		"TxtHost":     true,
		"TxtPort":     true,
		"TxtUsername": true,
		"TxtPassword": true,
		"TxtDbName":   true,
		"SelCharset":  false,
	}
)

type MainView struct {
	/* 基础控件 */
	App    fyne.App
	Window fyne.Window

	/* 控件 */
	// 连接控件
	SelDbType   *widget.Select
	TxtHost     *widget.Entry
	TxtPort     *widget.Entry
	TxtUsername *widget.Entry
	TxtPassword *widget.Entry
	TxtDbName   *widget.Entry
	SelCharset  *widget.Select

	// 输出控件
	TxtOutputDir       *widget.Entry
	BtnChooseOutputDir *widget.Button
	SelOutputFormat    *widget.Select

	// 按钮
	BtnTest     *widget.Button
	BtnGenerate *widget.Button

	// 容器
	container *fyne.Container

	/* 容器 */
	parentLayoutMap map[string]*fyne.Container
}

func NewMainView() *MainView {
	// 创建应用程序
	app := app.New()
	window := app.NewWindow("数据库连接配置")

	instance := &MainView{
		App:    app,
		Window: window,
	}

	// 初始化
	instance.init()

	return instance
}

func (this *MainView) initDefault() {
	// 设置默认输出目录
	if nil != this.TxtOutputDir && "" == this.TxtOutputDir.Text {
		// 用户桌面路径
		curPath := strings.TrimSpace(utils.GetUserDesktopPath())
		// 如果为空，设置当前目录
		if "" == curPath {
			curPath = "./"
		}
		// 更新
		this.TxtOutputDir.SetText(curPath)
	}

	// 调试模式下加载默认配置
	if !DEBUG {
		return
	}

	this.SelDbType.SetSelected("MySQL")
	this.TxtHost.SetText("www.example.com")
	this.TxtPort.SetText("3306")
	this.TxtUsername.SetText("root")
	this.TxtPassword.SetText("123456")
	this.TxtDbName.SetText("test")
	this.SelCharset.SetSelected("utf8mb4")
	this.SelOutputFormat.SetSelected("xlsx")
}
func (this *MainView) init() {
	// 设置窗口大小
	this.Window.Resize(fyne.NewSize(500, 400))
	this.Window.SetFixedSize(true)
	this.Window.RequestFocus()

	// 创建表单控件容器
	this.container = container.NewVBox(
		widget.NewLabelWithStyle("数据库连接配置", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
	)

	// 数据库类型
	this.SelDbType = widget.NewSelect(DbTypeList, this.selDbType_onChanged)
	this.SelDbType.SetSelectedIndex(0)
	// 数据库地址
	this.TxtHost = widget.NewEntry()
	this.TxtHost.SetPlaceHolder("例如: 192.168.1.100 或 SQLite 文件名 test.db")
	// 端口
	this.TxtPort = widget.NewEntry()
	this.TxtPort.SetPlaceHolder("例如: 3306")
	// 用户名
	this.TxtUsername = widget.NewEntry()
	this.TxtUsername.SetPlaceHolder("请输入用户名")
	// 密码
	this.TxtPassword = widget.NewPasswordEntry()
	this.TxtPassword.SetPlaceHolder("请输入密码")
	// 数据库名
	this.TxtDbName = widget.NewEntry()
	this.TxtDbName.SetPlaceHolder("请输入数据库名称")
	// 字符集
	this.SelCharset = widget.NewSelect(CharsetList, this.selCharset_onChanged)
	this.SelCharset.SetSelected("utf8mb4")

	// 输出目录
	this.TxtOutputDir = widget.NewEntry()
	this.TxtOutputDir.SetPlaceHolder("请指定输出目录")
	this.BtnChooseOutputDir = widget.NewButton("选择", this.btnChooseOutputDir_onClicked)
	outputDirContainer := container.NewBorder(nil, nil, nil, this.BtnChooseOutputDir, this.TxtOutputDir)

	// 输出格式
	this.SelOutputFormat = widget.NewSelect(OutputFormatList, this.selOutputFormat_onChanged)
	this.SelOutputFormat.SetSelected(OutputFormatList[0])

	// 测试按钮
	this.BtnTest = widget.NewButton("测试连接", this.btnTest_onClicked)
	// 生成文档按钮
	this.BtnGenerate = widget.NewButton("生成文档", this.btnGenerate_onClicked)
	// 按钮容器
	btnContainer := container.NewHBox(
		layout.NewSpacer(),
		this.BtnTest,
		layout.NewSpacer(),
		this.BtnGenerate,
		layout.NewSpacer(),
	)

	/* 添加到容器 */
	this.container.Add(this.createFormRow("数据库", "SelDbType", this.SelDbType))
	this.container.Add(layout.NewSpacer())
	this.container.Add(this.createFormRow("端口", "TxtPort", this.TxtPort))
	this.container.Add(this.createFormRow("地址", "TxtHost", this.TxtHost))
	this.container.Add(this.createFormRow("账号", "TxtUsername", this.TxtUsername))
	this.container.Add(this.createFormRow("密码", "TxtPassword", this.TxtPassword))
	this.container.Add(this.createFormRow("数据库名", "TxtDbName", this.TxtDbName))
	this.container.Add(this.createFormRow("字符集", "SelCharset", this.SelCharset))
	this.container.Add(layout.NewSpacer())
	this.container.Add(this.createFormRow("输出格式", "SelOutputFormat", this.SelOutputFormat))
	this.container.Add(this.createFormRow("输出目录", "SelOutputFormat", outputDirContainer))
	this.container.Add(layout.NewSpacer())
	this.container.Add(btnContainer)
	// 设置内边距容器
	rootContainer := container.NewPadded(this.container)

	// 添加到窗口
	this.Window.SetContent(rootContainer)

	// 默认值
	this.initDefault()
	// 切换显示模式
	this.changeDisplayMode(this.SelDbType.Selected)
	// 设置端口
	this.changePort(this.SelDbType.Selected)
}

// Show 显示窗口
func (this *MainView) Show() {
	this.Window.ShowAndRun()
}

func (this *MainView) validateForm() error {
	// 地址
	if "" == strings.TrimSpace(this.TxtHost.Text) {
		return errors.New("请填写数据库地址")
	}
	// 端口
	if !this.parentLayoutMap["TxtPort"].Hidden {
		p, err := strconv.Atoi(this.TxtPort.Text)
		if err != nil {
			return errors.New("无效的端口")
		}
		if 0 > p || 65535 < p {
			return errors.New("无效的端口")
		}
	}

	return nil
}

// createDatabaseConfig 创建数据库配置
func (this *MainView) createDatabaseConfig() (*configs.DatabaseConfig, error) {
	// 端口类型转换
	var port = 0
	if !this.parentLayoutMap["TxtPort"].Hidden {
		p, err := strconv.Atoi(this.TxtPort.Text)
		if err != nil {
			return nil, errors.New("无效的端口")
		}
		port = p
	}

	// 创建数据库配置
	dbConfig := &configs.DatabaseConfig{
		Type:     this.SelDbType.Selected,
		Host:     this.TxtHost.Text,
		Port:     port,
		Username: this.TxtUsername.Text,
		Password: this.TxtPassword.Text,
		Charset:  this.SelCharset.Selected,
		Database: this.TxtDbName.Text,
	}

	return dbConfig, nil
}

// createFormRow 创建表单行
func (this *MainView) createFormRow(text string, itemName string, itemWidget fyne.CanvasObject) *fyne.Container {
	lbl := widget.NewLabel(text)
	lbl.Alignment = fyne.TextAlignTrailing // 设置文本右对齐

	// 如果为空则创建，则创建
	if nil == this.parentLayoutMap {
		this.parentLayoutMap = make(map[string]*fyne.Container)
	}

	// 表单行
	formRow := container.New(layout.NewFormLayout(), lbl, itemWidget)
	this.parentLayoutMap[itemName] = formRow

	return formRow
}

// changeDisplayMode 修改显示模式
func (this *MainView) changeDisplayMode(selected string) {
	// 获取显示控件map
	displayModeMap, ok := DisplayModeMap[selected]
	if !ok {
		return
	}

	// 显示控件
	for itemName, itemVisible := range displayModeMap {
		// 找到控件
		reflectObj, _ := utils.ReflectFieldValue(this, itemName)
		if !reflectObj.IsValid() {
			return
		}

		// 找到上级容器
		if parentLayout, ok := this.parentLayoutMap[itemName]; ok {
			// 显示或隐藏
			if itemVisible {
				parentLayout.Show()
			} else {
				parentLayout.Hide()
			}
		}
	}
}

// changePort 端口选择变更事件处理函数
func (this *MainView) changePort(selected string) {
	// 默认端口
	port := DbPortMap[selected]
	// 设置端口
	if nil != this.TxtPort {
		this.TxtPort.SetText(fmt.Sprintf("%d", port))
	}
}

// ------------------------------
//
//	Event
//
// ------------------------------
// selDbType_onChanged 数据库类型下拉框选择变更事件处理函数
// 当用户在数据库类型下拉框中选择不同选项时触发此函数
// value: 用户选择的数据库类型值
func (this *MainView) selDbType_onChanged(value string) {
	// 修改显示模式
	this.changeDisplayMode(value)
	// 更新端口
	this.changePort(value)
}

// selCharset_onChanged 字符集下拉框选择变更事件处理函数
// 当用户在字符集下拉框中选择不同选项时触发此函数
// value: 用户选择的字符集值
func (this *MainView) selCharset_onChanged(value string) {
}

// selOutputFormat_onChanged 输出格式下拉框选择变更事件处理函数
// 当用户在输出格式下拉框中选择不同选项时触发此函数
// value: 用户选择的输出格式值
func (this *MainView) selOutputFormat_onChanged(value string) {
}

// btnChooseOutputDir_onClicked 选择输出目录按钮点击事件处理函数
func (this *MainView) btnChooseOutputDir_onClicked() {
	// 创建目录选择对话框
	dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, this.Window)
			return
		}
		if list == nil {
			return // 用户取消了选择
		}
		// 获取选择的目录路径
		this.TxtOutputDir.SetText(list.Path())
	}, this.Window)
}

// btnTest_onClicked 测试连接按钮点击事件处理函数
func (this *MainView) btnTest_onClicked() {
	// 创建数据库配置
	dbConfig, err := this.createDatabaseConfig()
	if nil != err {
		dialog.ShowError(fmt.Errorf("数据库配置错误: %w", err), this.Window)
		return
	}

	// 使用现有的 InitDatabase 方法初始化数据库连接
	db, err := configs.InitDatabase(dbConfig)
	if err != nil {
		dialog.ShowError(fmt.Errorf("数据库连接失败: %w", err), this.Window)
		return
	}

	// 执行简单的查询来测试连接是否正常
	sqlDB, err := db.DB()
	if err != nil {
		dialog.ShowError(fmt.Errorf("获取数据库连接池失败: %w", err), this.Window)
		return
	}

	// 尝试 ping 数据库
	if err = sqlDB.Ping(); err != nil {
		dialog.ShowError(fmt.Errorf("数据库 ping 失败: %w", err), this.Window)
		return
	}

	dialog.ShowInformation("提示", "数据库连接成功", this.Window)
}

// btnGenerate_onClicked 生成文档按钮点击事件处理函数
func (this *MainView) btnGenerate_onClicked() {
	// 初始化数据库配置
	dbConfig, err := this.createDatabaseConfig()
	if err != nil {
		dialog.ShowError(fmt.Errorf("初始化数据库配置失败: %w", err), this.Window)
		return
	}

	// 输出目录
	outputDirPath := this.TxtOutputDir.Text
	outputFormat := this.SelOutputFormat.Selected

	go func() {
		savePath, err := generateDict(dbConfig, outputDirPath, outputFormat)
		if err != nil {
			dialog.ShowError(fmt.Errorf("生成文档失败: %w", err), this.Window)
		} else {
			dialog.ShowInformation("提示", fmt.Sprintf("生成文档成功: %s", savePath), this.Window)
		}
	}()
}
