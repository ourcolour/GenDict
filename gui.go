package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"goDict/configs"
	"goDict/utils"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	// 数据库类型列表
	DbTypeList = []string{"MySQL", "SQLServer", "Oracle", "SQLite", "PostgreSQL"}
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
		// MySQL
		"MySQL": map[string]bool{
			"TxtHost":     true,
			"TxtPort":     true,
			"TxtUsername": true,
			"TxtPassword": true,
			"TxtDbName":   true,
			"SelCharset":  true,
		},
		// PostgreSQL
		"PostgreSQL": map[string]bool{
			"TxtHost":     true,
			"TxtPort":     true,
			"TxtUsername": true,
			"TxtPassword": true,
			"TxtDbName":   true,
			"SelCharset":  true,
		},
		// SQLServer
		"SQLServer": map[string]bool{
			"TxtHost":     true,
			"TxtPort":     true,
			"TxtUsername": true,
			"TxtPassword": true,
			"TxtDbName":   true,
			"SelCharset":  false,
		},
		// Oracle
		"Oracle": map[string]bool{
			"TxtHost":     true,
			"TxtPort":     true,
			"TxtUsername": true,
			"TxtPassword": true,
			"TxtDbName":   true,
			"SelCharset":  false,
		},
		// SQLite
		"SQLite": map[string]bool{
			"TxtHost":     true,
			"TxtPort":     false,
			"TxtUsername": false,
			"TxtPassword": false,
			"TxtDbName":   false,
			"SelCharset":  false,
		},
	}
)

type MainView struct {
	/* 基础控件 */
	App    fyne.App
	Window fyne.Window

	/* 表单 */
	// 基础表单
	FormBasic *widget.Form

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
	//parentLayoutMap map[string]*fyne.Container
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

// init 初始化
func (this *MainView) init() {
	/* 按钮 */
	this.BtnTest = widget.NewButtonWithIcon("测试连接", theme.DocumentCreateIcon(), this.btnTest_onClicked)
	this.BtnGenerate = widget.NewButtonWithIcon("现在生成", theme.BrokenImageIcon(), this.btnGenerate_onClicked)

	/* 表单控件 */
	// 数据库类型
	this.SelDbType = widget.NewSelect(DbTypeList, this.selDbType_onChanged)
	this.SelDbType.SetSelected(DbTypeList[0])
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
	//this.container.Add(this.createFormRow("输出目录", "SelOutputFormat", outputDirContainer))

	// 输出格式
	this.SelOutputFormat = widget.NewSelect(OutputFormatList, this.selOutputFormat_onChanged)
	this.SelOutputFormat.SetSelected(OutputFormatList[0])

	/* 表单 */
	// 基础表单
	this.FormBasic = &widget.Form{Items: []*widget.FormItem{
		widget.NewFormItem("数据库", this.SelDbType),
		widget.NewFormItem("地址", this.TxtHost),
		widget.NewFormItem("端口", this.TxtPort),
		widget.NewFormItem("账号", this.TxtUsername),
		widget.NewFormItem("密码", this.TxtPassword),
		widget.NewFormItem("库名", this.TxtDbName),
		widget.NewFormItem("字符集", this.SelCharset),
		widget.NewFormItem("输出格式", this.SelOutputFormat),
		widget.NewFormItem("输出目录", outputDirContainer),
	}}

	// 根容器
	rootContainer := container.NewPadded(container.NewVBox(
		&widget.Label{Text: "GenDict - 数据库字典生成工具", TextStyle: fyne.TextStyle{Bold: true, Italic: false, Monospace: false, Symbol: false, TabWidth: 0, Underline: false}, Alignment: 1, Wrapping: 0},
		widget.NewSeparator(),
		this.FormBasic,
		widget.NewSeparator(),
		container.NewGridWithColumns(3, layout.NewSpacer(), this.BtnTest, this.BtnGenerate),
	))

	this.Window = this.App.NewWindow("GenDict")
	this.Window.Resize(fyne.NewSize(460, 400))
	this.Window.SetContent(rootContainer)

	// 默认值
	this.initDefault()
	// 切换显示模式
	this.changeDisplayMode(this.SelDbType.Selected)
	// 设置端口
	this.changePort(this.SelDbType.Selected)
	// 调试模式自动填写
	this.initDebug("SQLite")
}

// initDefault 初始化默认值
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
}

func (this *MainView) initDebug(selected string) {
	// 调试模式下加载默认配置
	if !DEBUG {
		return
	}

	// 读取环境变量
	host := os.Getenv("host")
	port := os.Getenv("port")
	username := os.Getenv("username")
	password := os.Getenv("password")

	// 填写公共字段
	if "" == host {
		host = "localhost"
	}
	if "" == port {
		if "MySQL" == selected {
			username = "3306"
		} else if "SQLServer" == selected {
			username = "1433"
		} else if "Oracle" == selected {
			username = "1521"
		} else if "SQLite" == selected {
			username = "0"
		}
	}
	if "" == username {
		if "MySQL" == selected {
			username = "root"
		} else if "SQLServer" == selected {
			username = "sa"
		} else if "Oracle" == selected {
			username = "ora"
		} else if "SQLite" == selected {
			username = ""
		}
	}
	if "" == password {
		password = "123456"
	}
	this.TxtHost.SetText(host)
	this.TxtPort.SetText(port)
	this.TxtUsername.SetText(username)
	this.TxtPassword.SetText(password)

	this.SelDbType.SetSelected(selected)
	if "MySQL" == selected {
		this.TxtUsername.SetText(username)
		this.TxtPort.SetText(port)
		this.TxtDbName.SetText("student")
		this.SelCharset.SetSelected("utf8mb4")
		this.SelOutputFormat.SetSelected("xlsx")
	} else if "SQLServer" == selected {
		this.TxtUsername.SetText(username)
		this.TxtPort.SetText(port)
		this.TxtDbName.SetText("tel-qa")
		this.SelCharset.SetSelected("utf8mb4")
		this.SelOutputFormat.SetSelected("xlsx")
	} else if "Oracle" == selected {
		this.TxtUsername.SetText(username)
		this.TxtPort.SetText(port)
		this.TxtDbName.SetText("orcl")
		this.SelCharset.SetSelected("utf8mb4")
		this.SelOutputFormat.SetSelected("xlsx")
	} else if "SQLite" == selected {
		this.TxtPort.SetText(port)
		this.TxtDbName.SetText("")
		this.SelCharset.SetSelected("utf8mb4")
		this.SelOutputFormat.SetSelected("md")
	}
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
	if !this.TxtPort.Disabled() {
		p, err := strconv.Atoi(this.TxtPort.Text)
		if err != nil {
			return errors.New("无效的端口")
		}
		if 0 > p || 65535 < p {
			return errors.New("无效的端口")
		}
	}
	// 账号
	if !this.TxtUsername.Disabled() {
		if "" == strings.TrimSpace(this.TxtUsername.Text) {
			return errors.New("请填写账号")
		}
	}
	// 输出目录
	if !this.TxtOutputDir.Disabled() {
		if "" == strings.TrimSpace(this.TxtOutputDir.Text) {
			return errors.New("请填写输出目录")
		}
	}
	// 库名
	if !this.TxtDbName.Disabled() {
		if "" == strings.TrimSpace(this.TxtDbName.Text) {
			return errors.New("请填写数据库名称")
		}
	}

	return nil
}

// createDatabaseConfig 创建数据库配置
func (this *MainView) createDatabaseConfig() (*configs.DatabaseConfig, error) {
	// 端口类型转换
	var port = 0
	if !this.TxtPort.Disabled() {
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

// changeDisplayMode 修改显示模式
func (this *MainView) changeDisplayMode(selected string) {
	// 获取显示控件map
	displayModeMap, ok := DisplayModeMap[selected]
	if !ok {
		return
	}

	// 显示控件
	for controlName, controlEnabled := range displayModeMap {
		// 找到控件
		reflectObj, err := utils.ReflectFieldValue(this, controlName)
		if nil != err || !reflectObj.IsValid() {
			continue
		}

		// 获取对象
		pointer := reflectObj.Interface()
		// 完整的 nil 检查
		if pointer == nil || (reflect.ValueOf(pointer).Kind() == reflect.Ptr && reflect.ValueOf(pointer).IsNil()) {
			continue
		}

		// 直接使用类型断言获取 fyne.CanvasObject 接口
		if control, ok := pointer.(fyne.Disableable); ok {
			// 显示或隐藏
			if controlEnabled {
				control.Enable()
			} else {
				control.Disable()
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

	// 验证表单
	err = this.validateForm()
	if err != nil {
		dialog.ShowError(err, this.Window)
		return
	}

	go func() {
		savePath, err := generateDict(dbConfig, outputDirPath, outputFormat)
		if err != nil {
			dialog.ShowError(err, this.Window)
		} else {
			dialog.ShowInformation("提示", fmt.Sprintf("生成文档成功: %s", savePath), this.Window)
		}
	}()
}
