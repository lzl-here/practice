package mapping

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

func getValueByPath(obj interface{}, path string) (interface{}, error) {
	current := reflect.ValueOf(obj)
	if current.Kind() == reflect.Ptr {
		current = current.Elem() 
	}

	segments := strings.Split(path, ".")
	for _, seg := range segments {
		if current.Kind() != reflect.Struct {
			return nil, fmt.Errorf("invalid path segment %s for type %s", seg, current.Type())
		}

		fieldFound := false
		// 遍历所有字段进行大小写不敏感匹配
		for i := 0; i < current.NumField(); i++ {
			field := current.Type().Field(i)
			if strings.EqualFold(field.Name, seg) {
				current = current.Field(i)
				fieldFound = true
				break
			}
		}

		if !fieldFound {
			return nil, fmt.Errorf("field %s not found in struct %s", seg, current.Type())
		}

		// 处理嵌套指针
		if current.Kind() == reflect.Ptr {
			if current.IsNil() {
				return nil, nil // 允许空指针返回nil
			}
			current = current.Elem()
		}
	}
	return current.Interface(), nil
}

func loadFromConfig(configPath string) ([]*MappingRule, error) {
	// 读取文件内容
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析 YAML 文件
	var rules []*MappingRule
	err = yaml.Unmarshal(data, &rules)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return rules, nil
}
