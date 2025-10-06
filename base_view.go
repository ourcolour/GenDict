package main

import (
	"goDict/utils"
)

// 函数别名
var (
	I  = utils.I18nGetTextByMessageId
	Id = utils.I18nGetTextByMessageIdAndTemplateData
)

// BaseView 基础视图
type BaseView struct {
}

// I18nGetTextByMessageId 获取i18n文本
func (bv *BaseView) I(messageId string) string {
	return utils.I18nGetTextByMessageId(messageId)
}

// I18nGetTextByMessageIdAndTemplateData 获取i18n文本
func (bv *BaseView) Id(messageId string, data map[string]interface{}) string {
	return utils.I18nGetTextByMessageIdAndTemplateData(messageId, data)
}
