package mapping

// 定义映射规则配置结构
type MappingRule struct {
	SourcePath string `yaml:"source_path"` // 源字段路径，如："User.Profile.Age"
	TargetKey  string `yaml:"target_key"`  // 目标参数名，如："user_age"
	Transform  string `yaml:"transform"`   // 值转换表达式，如："value > 18 ? 'adult' : 'child'"
	Condition  string `yaml:"condition"`   // 应用条件，如："order.Country == 'CN'"
}

// 转换函数注册表
type TransformFunc func(interface{}) interface{}

var transformRegistry = make(map[string]TransformFunc)

// 条件判断引擎接口
type ConditionEvaluator interface {
	Evaluate(condition string, params map[string]interface{}) (bool, error)
}

type EngineImpl struct{

}

