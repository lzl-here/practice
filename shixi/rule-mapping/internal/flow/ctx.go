package flow

import (
	"reflect"
	"rule-mapping/internal/model"
)

type FlowContext struct {
	Order *model.Order
}

func (contx *FlowContext) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	ctxValue := reflect.ValueOf(contx).Elem() // 解引用指针接收者
	ctxType := ctxValue.Type()

	for i := 0; i < ctxValue.NumField(); i++ {
		field := ctxType.Field(i)
		fieldValue := ctxValue.Field(i)
		result[field.Name] = fieldValue.Interface() // 自动处理指针类型
	}
	return result
}