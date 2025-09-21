package utils

import (
	"fmt"
	"reflect"
)

func ReflectFieldValue(obj interface{}, fieldName string) (*reflect.Value, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	value := val.FieldByName(fieldName)
	if !value.IsValid() {
		return nil, fmt.Errorf("字段 %s 不存在", fieldName)
	}

	return &value, nil
}
