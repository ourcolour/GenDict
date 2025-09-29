package utils

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	//go:embed i18n/locale.*.json
	LocaleFS embed.FS

	// I18n 文件格式
	i18nFormat = "json"
	// I18n 默认语言
	i18nDefaultLocale = language.English
	// 当前语言标签
	i18nCurrentLocale = i18nDefaultLocale
	// I18n 语言方案
	i18nLocaleArray = []string{"en", "zh"}
	// I18n Bundle
	i18nBundle *i18n.Bundle
	// I18n Localizer
	i18nLocalizer *i18n.Localizer
	// I18n 消息文件
	i18nMessageFile *i18n.MessageFile
)

// InitI18n 初始化i18n
func InitI18n() error {
	// 初始化 i18n bundle
	i18nBundle = i18n.NewBundle(i18nDefaultLocale)
	i18nBundle.RegisterUnmarshalFunc(i18nFormat, json.Unmarshal)

	// 加载语言
	return I18nLoadLanguage(i18nDefaultLocale)
}

// I18nGetTextByMessageId 获取i18n文本
func I18nGetTextByMessageId(msgId string) string {
	return I18nGetTextByMessageIdAndTemplateData(msgId, nil)
}

// I18nGetTextByMessageId 获取i18n文本
func I18nGetTextByMessageIdAndTemplateData(msgId string, templateData map[string]interface{}) string {
	text := i18nLocalizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    msgId,
		TemplateData: templateData,
	})

	return text
}

// I18nLoadLanguage 加载语言
func I18nLoadLanguage(tag language.Tag) error {
	// 解析语言标签
	if language.Und == tag {
		return fmt.Errorf("Invalid locale: %s", tag.String())
	}

	// 更新当前语言
	i18nCurrentLocale = tag

	// 重新加载消息文件
	_, err := i18nBundle.LoadMessageFileFS(LocaleFS, fmt.Sprintf("i18n/locale.%s.json", tag.String()))
	if err != nil {
		return err
	}
	// 创建新的Localizer
	i18nLocalizer = i18n.NewLocalizer(i18nBundle, tag.String())

	return nil
}

// I18nGetCurrentLocale 获取当前语言标签
func I18nGetCurrentLocale() string {
	return i18nCurrentLocale.String()
}

// I18nGetAvailableLocales 获取可用语言列表
func I18nGetAvailableLocales() []string {
	return i18nLocaleArray
}
