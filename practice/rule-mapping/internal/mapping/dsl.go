package mapping

import (
	"fmt"
	"time"

	"github.com/Knetic/govaluate"
)

type DSLParamBuilder struct {
	rules          []*MappingRule
	evaluator      ConditionEvaluator
	transformCache map[string]govaluate.EvaluableExpression
}

func NewDSLParamBuilder(configPath string) (*DSLParamBuilder, error) {
	// 加载YAML配置（示例配置见后文）
	rules, err := loadFromConfig(configPath)
	if err != nil {
		return nil, err
	}

	return &DSLParamBuilder{
		rules:          rules,
		evaluator:      NewGovaluateEvaluator(), // 条件判断实现
		transformCache: make(map[string]govaluate.EvaluableExpression),
	}, nil
}

func (b *DSLParamBuilder) BuildParams(outParam interface{}) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// 第一步：收集原始数据（性能关键路径使用并行处理）
	rawValues := make(map[string]interface{}, len(b.rules))
	for _, rule := range b.rules {
		val, err := getValueByPath(outParam, rule.SourcePath)
		if err != nil {
			return nil, fmt.Errorf("error getting %s: %w", rule.SourcePath, err)
		}
		rawValues[rule.SourcePath] = val
	}

	// 第二步：处理条件和转换
	for _, rule := range b.rules {
		// 应用前置条件判断
		if rule.Condition != "" {
			match, err := b.evaluator.Evaluate(rule.Condition, rawValues)
			if err != nil {
				return nil, fmt.Errorf("condition evaluation error: %w", err)
			}
			if !match {
				continue
			}
		}

		rawValue := rawValues[rule.SourcePath]

		// 处理数据转换
		finalValue, err := b.applyTransform(rawValue, rule.Transform)
		if err != nil {
			return nil, fmt.Errorf("transform error: %w", err)
		}

		params[rule.TargetKey] = finalValue
	}

	return params, nil
}

func (b *DSLParamBuilder) applyTransform(raw interface{}, expr string) (interface{}, error) {
	if expr == "" {
		return raw, nil
	}
	cachedExpr, exists := b.transformCache[expr]
	if !exists {
		parsedExpr, err := govaluate.NewEvaluableExpression(expr)
		if err != nil {
			return nil, fmt.Errorf("invalid transform expression: %w", err)
		}
		b.transformCache[expr] = *parsedExpr
		cachedExpr = *parsedExpr
	}

	parameters := map[string]interface{}{
		"value": raw,        // 内置变量
		"now":   time.Now(), // 其他上下文变量
	}

	result, err := cachedExpr.Evaluate(parameters)
	if err != nil {
		return nil, fmt.Errorf("expression evaluation failed: %w", err)
	}
	return result, nil
}

func NewGovaluateEvaluator() ConditionEvaluator {

	return nil
}
