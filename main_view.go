package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"goDict/configs"
	"goDict/utils"
	"golang.org/x/text/language"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	// 数据库类型列表
	DbTypeList = []string{"MySQL", "SQLServer", "Oracle", "SQLite", "PostgresSQL"}
	// 数据库端口
	DbPortMap = map[string]int{
		"MySQL":       3306,
		"SQLServer":   1433,
		"PostgresSQL": 5432,
		"Oracle":      1521,
		"SQLite":      0,
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
			"TxtService":  false,
		},
		// PostgresSQL
		"PostgresSQL": map[string]bool{
			"TxtHost":     true,
			"TxtPort":     true,
			"TxtUsername": true,
			"TxtPassword": true,
			"TxtDbName":   true,
			"SelCharset":  true,
			"TxtService":  false,
		},
		// SQLServer
		"SQLServer": map[string]bool{
			"TxtHost":     true,
			"TxtPort":     true,
			"TxtUsername": true,
			"TxtPassword": true,
			"TxtDbName":   true,
			"SelCharset":  false,
			"TxtService":  false,
		},
		// Oracle
		"Oracle": map[string]bool{
			"TxtHost":     true,
			"TxtPort":     true,
			"TxtUsername": true,
			"TxtPassword": true,
			"TxtDbName":   true,
			"SelCharset":  false,
			"TxtService":  true,
		},
		// SQLite
		"SQLite": map[string]bool{
			"TxtHost":     true,
			"TxtPort":     false,
			"TxtUsername": false,
			"TxtPassword": false,
			"TxtDbName":   false,
			"SelCharset":  false,
			"TxtService":  false,
		},
	}
)

type MainView struct {
	*BaseView

	/* 基础控件 */
	App    *fyne.App
	Window *fyne.Window

	/* 表单 */
	// 标题
	LblTitle *widget.Label
	// 基础表单
	FormBasic *widget.Form

	/* 控件 */
	// 连接控件
	SelDbType *widget.Select
	TxtHost   *widget.Entry
	TxtPort   *widget.Entry
	// Oracle SID/Service
	TxtService  *widget.Entry
	TxtUsername *widget.Entry
	TxtPassword *widget.Entry
	TxtDbName   *widget.Entry
	SelCharset  *widget.Select

	// 输出控件
	TxtOutputDir       *widget.Entry
	BtnChooseOutputDir *widget.Button
	SelOutputFormat    *widget.Select

	// 语言选择控件
	SelLocale *widget.Select
	// 按钮
	BtnTest              *widget.Button
	BtnGenerate          *widget.Button
	BtnCustomizeGenerate *widget.Button

	// 容器
	container *fyne.Container
}

func NewMainView(appInstance *fyne.App) *MainView {
	// 创建应用程序
	instance := &MainView{}

	// 初始化
	instance.init(appInstance)

	return instance
}

// init 初始化
func (this *MainView) init(appInstance *fyne.App) {
	// 应用
	this.App = appInstance
	// 窗口
	window := (*this.App).NewWindow(I("main-view.ui.window.title"))
	this.Window = &window

	/* 标题 */
	this.LblTitle = &widget.Label{
		Text:      I("main-view.ui.rootContainer.label.text"),
		TextStyle: fyne.TextStyle{Bold: true, Italic: false, Monospace: false, Symbol: false, TabWidth: 0, Underline: false},
		Alignment: 1,
		Wrapping:  0,
	}

	/* 按钮 */
	this.BtnTest = widget.NewButtonWithIcon(I("main-view.ui.BtnTest.label"), theme.InfoIcon(), this.btnTest_onClicked)
	this.BtnCustomizeGenerate = widget.NewButtonWithIcon(I("main-view.ui.BtnCustomizeGenerate.label"), theme.MediaSkipNextIcon(), this.btnCustomizeGenerate_onClicked)
	this.BtnGenerate = widget.NewButtonWithIcon(I("main-view.ui.BtnGenerate.label"), theme.MediaPlayIcon(), this.btnGenerate_onClicked)
	this.BtnGenerate.Importance = widget.HighImportance

	/* 表单控件 */
	// 数据库类型
	this.SelDbType = widget.NewSelect(DbTypeList, this.selDbType_onChanged)
	this.SelDbType.SetSelected(DbTypeList[0])
	// 数据库地址
	this.TxtHost = widget.NewEntry()
	this.TxtHost.SetPlaceHolder(I("main-view.ui.TxtHost.placeholder"))
	// 端口
	this.TxtPort = widget.NewEntry()
	this.TxtPort.SetPlaceHolder(I("main-view.ui.TxtPort.placeholder"))
	// 用户名
	this.TxtUsername = widget.NewEntry()
	this.TxtUsername.SetPlaceHolder(I("main-view.ui.TxtUsername.placeholder"))
	// Oracle SID/Service
	this.TxtService = widget.NewEntry()
	this.TxtService.SetPlaceHolder(I("main-view.ui.TxtService.placeholder"))
	// 密码
	this.TxtPassword = widget.NewPasswordEntry()
	this.TxtPassword.SetPlaceHolder(I("main-view.ui.TxtPassword.placeholder"))
	// 数据库名
	this.TxtDbName = widget.NewEntry()
	this.TxtDbName.SetPlaceHolder(I("main-view.ui.TxtDbName.placeholder"))
	// 字符集
	this.SelCharset = widget.NewSelect(CharsetList, this.selCharset_onChanged)
	this.SelCharset.SetSelected("utf8mb4")

	// 输出目录
	this.TxtOutputDir = widget.NewEntry()
	this.TxtOutputDir.SetPlaceHolder(I("main-view.ui.TxtOutputDir.placeholder"))
	this.BtnChooseOutputDir = widget.NewButton(I("main-view.ui.BtnChooseOutputDir.placeholder"), this.btnChooseOutputDir_onClicked)
	outputDirContainer := container.NewBorder(nil, nil, nil, this.BtnChooseOutputDir, this.TxtOutputDir)

	// 输出格式
	this.SelOutputFormat = widget.NewSelect(OutputFormatList, this.selOutputFormat_onChanged)
	this.SelOutputFormat.SetSelected(OutputFormatList[0])

	/* 表单 */
	// 基础表单
	this.FormBasic = &widget.Form{Items: []*widget.FormItem{
		widget.NewFormItem(I("main-view.ui.form.formItem.SelDbType.text"), this.SelDbType),
		widget.NewFormItem(I("main-view.ui.form.formItem.TxtHost.text"), this.TxtHost),
		widget.NewFormItem(I("main-view.ui.form.formItem.TxtPort.text"), this.TxtPort),
		widget.NewFormItem(I("main-view.ui.form.formItem.TxtService.text"), this.TxtService),
		widget.NewFormItem(I("main-view.ui.form.formItem.TxtUsername.text"), this.TxtUsername),
		widget.NewFormItem(I("main-view.ui.form.formItem.TxtPassword.text"), this.TxtPassword),
		widget.NewFormItem(I("main-view.ui.form.formItem.TxtDbName.text"), this.TxtDbName),
		widget.NewFormItem(I("main-view.ui.form.formItem.SelCharset.text"), this.SelCharset),
		widget.NewFormItem(I("main-view.ui.form.formItem.SelOutputFormat.text"), this.SelOutputFormat),
		widget.NewFormItem(I("main-view.ui.form.formItem.outputDirContainer.text"), outputDirContainer),
	}}

	/* 语言选择 */
	this.SelLocale = widget.NewSelect(utils.I18nGetAvailableLocales(), this.selLocale_onChanged)
	this.SelLocale.SetSelected(utils.I18nGetCurrentLocale())

	// 根容器
	rootContainer := container.NewPadded(container.NewVBox(
		container.NewHBox(this.LblTitle, layout.NewSpacer(), this.SelLocale),
		widget.NewSeparator(),
		this.FormBasic,
		widget.NewSeparator(),
		container.NewBorder(nil, nil, this.BtnTest, container.NewHBox(this.BtnCustomizeGenerate, this.BtnGenerate), nil),
	))

	// 窗口尺寸
	(*this.Window).Resize(fyne.NewSize(580, 400))
	(*this.Window).SetContent(rootContainer)

	// 默认值
	this.initDefault()
	// 切换显示模式
	this.changeDisplayMode(this.SelDbType.Selected)
	// 设置端口
	this.changePort(this.SelDbType.Selected)
	// 调试模式自动填写
	this.initDebug("MySQL")
}

// refreshUITexts 重新翻译界面文本
func (this *MainView) refreshUILocale() {
	// 更新窗口标题
	(*this.Window).SetTitle(I("main-view.ui.window.title"))
	// 窗口内标题标签
	this.LblTitle.SetText(I("main-view.ui.rootContainer.label.text"))

	// 更新按钮文本
	this.BtnTest.SetText(I("main-view.ui.BtnTest.label"))
	this.BtnGenerate.SetText(I("main-view.ui.BtnGenerate.label"))
	this.BtnCustomizeGenerate.SetText(I("main-view.ui.BtnCustomizeGenerate.label"))

	// 更新占位符文本
	this.TxtHost.SetPlaceHolder(I("main-view.ui.TxtHost.placeholder"))
	this.TxtPort.SetPlaceHolder(I("main-view.ui.TxtPort.placeholder"))
	this.TxtService.SetPlaceHolder(I("main-view.ui.TxtService.placeholder"))
	this.TxtUsername.SetPlaceHolder(I("main-view.ui.TxtUsername.placeholder"))
	this.TxtPassword.SetPlaceHolder(I("main-view.ui.TxtPassword.placeholder"))
	this.TxtDbName.SetPlaceHolder(I("main-view.ui.TxtDbName.placeholder"))
	this.TxtOutputDir.SetPlaceHolder(I("main-view.ui.TxtOutputDir.placeholder"))
	this.BtnChooseOutputDir.SetText(I("main-view.ui.BtnChooseOutputDir.placeholder"))

	// 更新表单项标签
	if len(this.FormBasic.Items) >= 8 {
		this.FormBasic.Items[0].Text = I("main-view.ui.form.formItem.SelDbType.text")
		this.FormBasic.Items[1].Text = I("main-view.ui.form.formItem.TxtHost.text")
		this.FormBasic.Items[2].Text = I("main-view.ui.form.formItem.TxtPort.text")
		this.FormBasic.Items[3].Text = I("main-view.ui.form.formItem.TxtService.text")
		this.FormBasic.Items[4].Text = I("main-view.ui.form.formItem.TxtUsername.text")
		this.FormBasic.Items[5].Text = I("main-view.ui.form.formItem.TxtPassword.text")
		this.FormBasic.Items[6].Text = I("main-view.ui.form.formItem.TxtDbName.text")
		this.FormBasic.Items[7].Text = I("main-view.ui.form.formItem.SelCharset.text")
		this.FormBasic.Items[8].Text = I("main-view.ui.form.formItem.SelOutputFormat.text")
		this.FormBasic.Items[9].Text = I("main-view.ui.form.formItem.outputDirContainer.text")
		this.FormBasic.Refresh()
	}

	// 更新标签文本
	// 如果有其他标签也需要更新，可以在这里添加
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
			port = "3306"
		} else if "PostgresSQL" == selected {
			port = "5432"
		} else if "SQLServer" == selected {
			port = "1433"
		} else if "Oracle" == selected {
			port = "1521"
		} else if "SQLite" == selected {
			port = "0"
		}
	}
	if "" == username {
		if "MySQL" == selected {
			username = "root"
		} else if "PostgresSQL" == selected {
			username = "postgres"
		} else if "SQLServer" == selected {
			username = "sa"
		} else if "Oracle" == selected {
			username = "c##cc"
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
	} else if "PostgresSQL" == selected {
		this.TxtUsername.SetText(username)
		this.TxtPort.SetText(port)
		this.TxtDbName.SetText("album_web")
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
		this.TxtService.SetText("ORCLCDB")
		this.TxtDbName.SetText("c##cc")
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
	(*this.Window).SetOnClosed(func() {
		os.Exit(0) // 窗口关闭时退出程序
	})
	(*this.Window).ShowAndRun()
}

func (this *MainView) validateForm() error {
	// 地址
	if "" == strings.TrimSpace(this.TxtHost.Text) {
		return errors.New(I("main-view.msg.error.databaseAddressRequired"))
	}
	// 端口
	if !this.TxtPort.Disabled() {
		p, err := strconv.Atoi(this.TxtPort.Text)
		if err != nil {
			return errors.New(I("main-view.msg.error.invalidPort"))
		}
		if 0 > p || 65535 < p {
			return errors.New(I("main-view.msg.error.invalidPort"))
		}
	}
	// 账号
	if !this.TxtUsername.Disabled() {
		if "" == strings.TrimSpace(this.TxtUsername.Text) {
			return errors.New(I("main-view.msg.error.usernameRequired"))
		}
	}
	// 输出目录
	if !this.TxtOutputDir.Disabled() {
		if "" == strings.TrimSpace(this.TxtOutputDir.Text) {
			return errors.New(I("main-view.msg.error.outputDirRequired"))
		}
	}
	// 库名
	if !this.TxtDbName.Disabled() {
		if "" == strings.TrimSpace(this.TxtDbName.Text) {
			return errors.New(I("main-view.msg.error.databaseNameRequired"))
		}
	}
	// Oracle启用Service
	if "Oracle" == this.SelDbType.Selected {
		if "" == strings.TrimSpace(this.TxtService.Text) {
			return errors.New(I("main-view.msg.error.serviceRequired"))
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
			return nil, errors.New(I("main-view.msg.error.invalidPort"))
		}
		port = p
	}

	// 创建数据库配置
	dbConfig := &configs.DatabaseConfig{
		Type:     this.SelDbType.Selected,
		Host:     this.TxtHost.Text,
		Port:     port,
		Service:  this.TxtService.Text,
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

	// 属性
	if nil != this.FormBasic {
		this.FormBasic.Refresh()
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
			dialog.ShowError(errors.New(I("main-view.msg.error.folderOpenError")), (*this.Window))
			return
		}
		if list == nil {
			return // 用户取消了选择
		}
		// 获取选择的目录路径
		this.TxtOutputDir.SetText(list.Path())
	}, (*this.Window))
}

// selLocale_onChanged 选择语言下拉框选择变更事件处理函数
func (this *MainView) selLocale_onChanged(selected string) {
	// 加载语言标签
	tag := language.Make(selected)
	// 加载新语言
	err := utils.I18nLoadLanguage(tag)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load language: %v", err), (*this.Window))
		return
	}

	// 重新翻译界面元素
	this.refreshUILocale()
}

// btnTest_onClicked 测试连接按钮点击事件处理函数
func (this *MainView) btnTest_onClicked() {
	// 创建数据库配置
	dbConfig, err := this.createDatabaseConfig()
	if nil != err {
		dialog.ShowError(errors.New(Id("main-view.msg.error.invalidDatabaseConfig", map[string]interface{}{"Error": err.Error()})), (*this.Window))
		return
	}

	// 使用现有的 InitDatabase 方法初始化数据库连接
	db, err := configs.InitDatabase(dbConfig)
	if err != nil {
		dialog.ShowError(errors.New(Id("main-view.msg.error.databaseConnectingFailed", map[string]interface{}{"Error": err.Error()})), (*this.Window))
		return
	}

	// 执行简单的查询来测试连接是否正常
	sqlDB, err := db.DB()
	if err != nil {
		dialog.ShowError(errors.New(Id("main-view.msg.error.databaseFetchingConnectionPoolFailed", map[string]interface{}{"Error": err.Error()})), (*this.Window))
		return
	}

	// 尝试 ping 数据库
	if err = sqlDB.Ping(); err != nil {
		dialog.ShowError(errors.New(Id("main-view.msg.error.databasePingingFailed", map[string]interface{}{"Error": err.Error()})), (*this.Window))
		return
	}

	dialog.ShowInformation(
		I("main-view.msg.info"),
		I("main-view.msg.error.databaseConnectingSucceeded"),
		(*this.Window),
	)
}

// btnCustomizeGenerate_onClicked 自定义生成按钮点击事件处理函数
func (this *MainView) btnCustomizeGenerate_onClicked() {
	//dialog.ShowInformation("Info", "User customized generation coming soon ...", (*this.Window))
	//return

	// 初始化数据库配置
	dbConfig, err := this.createDatabaseConfig()
	if err != nil {
		dialog.ShowError(errors.New(Id("main-view.msg.error.databaseInitializingFailed", map[string]interface{}{"Error": err.Error()})), (*this.Window))
		return
	}

	// 输出目录
	outputDirPath := this.TxtOutputDir.Text
	outputFormat := this.SelOutputFormat.Selected

	// 验证表单
	err = this.validateForm()
	if err != nil {
		dialog.ShowError(errors.New(I("main-view.msg.error.validateFormError")), (*this.Window))
		return
	}

	// 获取数据库信息
	dbInfo, err := getDatabaseInfo(dbConfig, nil)
	if err != nil {
		dialog.ShowError(errors.New(Id("main-view.msg.error.databaseInfoFetchingFailed", map[string]interface{}{"Error": err.Error()})), (*this.Window))
		return
	}

	// 提取数据库信息
	result := make(map[string]string, len(dbInfo.TableNameList))
	for _, tableInfo := range dbInfo.TableMap {
		result[tableInfo.TableName] = tableInfo.TableType
	}

	// 显示选择窗口
	searchView := NewSearchView(this.App, result)
	// 注册回调事件
	searchView.OnFinished = func(selectedTableNameList []string) {
		slog.Debug("selectedMap: %+v", selectedTableNameList)

		go func() {
			fyne.Do(func() {
				// 主窗口获得焦点
				(*this.Window).RequestFocus()
			})

			// 生成数据
			savePath, err := generateDict(dbConfig, outputDirPath, outputFormat, selectedTableNameList)
			if err != nil {
				dialog.ShowError(errors.New(Id("main-view.msg.error.generateDictError", map[string]interface{}{"Error": err.Error()})), (*this.Window))
				return
			}

			dialog.ShowInformation(
				I("main-view.msg.info"),
				Id("main-view.msg.error.documentCreatingSucceeded", map[string]interface{}{"Message": savePath}),
				(*this.Window),
			)
		}()
	}
	// 显示选择窗口
	searchView.Show()
}

// btnGenerate_onClicked 生成文档按钮点击事件处理函数
func (this *MainView) btnGenerate_onClicked() {
	// 初始化数据库配置
	dbConfig, err := this.createDatabaseConfig()
	if err != nil {
		dialog.ShowError(errors.New(Id("main-view.msg.error.databaseInitializingFailed", map[string]interface{}{"Error": err.Error()})), (*this.Window))
		return
	}

	// 输出目录
	outputDirPath := this.TxtOutputDir.Text
	outputFormat := this.SelOutputFormat.Selected

	// 验证表单
	err = this.validateForm()
	if err != nil {
		dialog.ShowError(errors.New(I("main-view.msg.error.validateFormError")), (*this.Window))
		return
	}

	go func() {
		// 生成
		savePath, err := generateDict(dbConfig, outputDirPath, outputFormat, nil)
		if err != nil {
			dialog.ShowError(errors.New(Id("main-view.msg.error.generateDictError", map[string]interface{}{"Error": err.Error()})), (*this.Window))
			return
		}

		dialog.ShowInformation(
			I("main-view.msg.info"),
			Id("main-view.msg.error.documentCreatingSucceeded", map[string]interface{}{"Message": savePath}),
			(*this.Window),
		)
	}()
}
