// services/GeneratorService.go
package services

import (
	"errors"
	"fmt"
	"gorm.io/gen"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"
)

type TemplateData struct {
	SnakeModelName string
	UpperModelName string
	LowerModelName string
	ModelDirPath   string
	ServiceDirPath string
}

func NewTemplateData(snakeModelName string, upperModelName string, lowerModelName string, modelDirPathString string, serviceDirPathString string) *TemplateData {
	return &TemplateData{
		SnakeModelName: snakeModelName,
		UpperModelName: upperModelName,
		LowerModelName: lowerModelName,
		ModelDirPath:   modelDirPathString,
		ServiceDirPath: serviceDirPathString,
	}
}

func (this *TemplateData) GetModelFilePath() string {
	return path.Join(this.ModelDirPath, fmt.Sprintf("%s%s", this.SnakeModelName, ".gen.go"))
}
func (this *TemplateData) GetServiceFilePath() string {
	return path.Join(this.ModelDirPath, fmt.Sprintf("%s%s", this.UpperModelName, "Service.gen.go"))
}

type GeneratorService struct {
	db *gorm.DB
}

func NewGeneratorService(db *gorm.DB) *GeneratorService {
	return &GeneratorService{
		db: db,
	}
}

func (gs *GeneratorService) GenerateModels(outPath, modelDirPath string, serviceDirPath string) (modelArray []interface{}) {
	// 初始化GORM Gen生成器
	g := gen.NewGenerator(gen.Config{
		OutPath:       outPath,
		ModelPkgPath:  modelDirPath,
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable: true,
	})

	g.UseDB(gs.db)

	// 为每个表生成模型配置
	modelArray = g.GenerateAllTable()

	// 执行生成
	g.Execute()

	// 生成模板
	if 1 > len(modelArray) {
		return modelArray
	}

	// 志哥生成模板文件
	for _, model := range modelArray {
		// 获取 model 的实际值（通过反射）
		v := reflect.ValueOf(model).Elem()
		if v.Kind() == reflect.Struct {
			// 尝试获取 fileName 字段
			fileNameField := v.FieldByName("FileName")
			modelStructField := v.FieldByName("ModelStructName")
			queryStructField := v.FieldByName("QueryStructName")

			if !gs.isValidField(fileNameField, modelStructField, queryStructField) {
				continue
			}

			// 创建模板数据
			templateData := NewTemplateData(fileNameField.String(), modelStructField.String(), queryStructField.String(), modelDirPath, serviceDirPath)

			// 处理 model
			gs.postProcessModel(templateData)

			// 根据模板生成service
			serviceFilePath := gs.GenerateService(templateData)
			slog.Info("生成services文件名: " + serviceFilePath)
		}
	}

	return modelArray
}

// isValidField 判断字段是否有效
func (gs *GeneratorService) isValidField(fieldArray ...reflect.Value) bool {
	// 遍历每一个字段
	for _, field := range fieldArray {
		cur := field.IsValid() && field.Kind() == reflect.String
		if !cur {
			return false
		}
	}
	return true
}

// postProcessModel 处理model文件
func (gs *GeneratorService) postProcessModel(templateData *TemplateData) (string, error) {
	// Args
	if nil == templateData {
		return "", errors.New("无效的模板数据")
	}

	// model文件路径
	modelPath := templateData.GetModelFilePath()
	// 获取文件权限
	info, err := os.Stat(modelPath)
	if err != nil {
		return "", err
	}
	// 加载生成的model文件
	file, err := os.Open(modelPath)
	// 文件权限
	fileMode := info.Mode().Perm()
	defer file.Close()

	// 读取全部内容
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// 处理换行符
	content := strings.ReplaceAll(strings.ReplaceAll(string(fileContent), "\r", "\n"), "\n\n", "\n")

	// 转为行
	lineArray := strings.Split(content, "\n")
	newLineArray := []string{}
	// 遍历行
	for _, line := range lineArray {
		// 忽略空行
		if 1 > len(strings.TrimSpace(line)) {
			continue
		}

		// 在TableName()方法处，改为值引用而不是指针
		if strings.Contains(line, " TableName() string") {
			line = strings.ReplaceAll(line, "func (*", "func (")
		}
		// 添加当前行
		newLineArray = append(newLineArray, line)

		// 在 type 后添继承
		if strings.HasPrefix(line, "type") {
			newLineArray = append(newLineArray, "\tBaseModel")
		}
	}
	// 更新内容
	content = strings.Join(newLineArray, "\n")

	// 写入文件
	err = os.WriteFile(modelPath, []byte(content), fileMode)
	if err != nil {
		return "", err
	}

	return content, nil
}

func (gs *GeneratorService) GenerateService(templateData *TemplateData) string {
	// 模板数据
	templateDataString := map[string]string{
		// 变量名从templateData改为data，以匹配Execute方法
		"UpperModelName": templateData.UpperModelName,
		"lowerModelName": templateData.LowerModelName,
	}

	// 1. 打开模板文件
	fileContent, err := os.ReadFile("./resources/templates/gen_template.go")
	if err != nil {
		panic(err)
	}

	// 2. 转换为string
	content := strings.ReplaceAll(strings.ReplaceAll(string(fileContent), "\r", "\n"), "\n\n", "\n")

	// 3. 跳过前两行（包声明和导入）
	lines := strings.Split(content, "\n")
	// 从第3行开始（索引为2）拼接剩余内容
	filteredContent := strings.Join(lines[2:], "\n")

	// 4. 创建模板并解析
	t := template.Must(template.New("gen_template").Parse(filteredContent))

	// 5. 创建输出文件
	savePath := fmt.Sprintf("%s/%s_service.gen.go", templateData.ServiceDirPath, templateData.SnakeModelName)
	outFile, err := os.Create(savePath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// 6. 执行模板生成
	err = t.Execute(outFile, templateDataString)
	if err != nil {
		panic(err)
	}

	return savePath
}
