package rule

import (
	"fmt"
	"log"
	"os"
	"rule-mapping/internal/model"
	"strings"

	"github.com/Knetic/govaluate"
	"gopkg.in/yaml.v3"
)

func LoadFromConfig(configPath string) (*RuleTree, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML配置
	var cfg TreeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 构建节点映射表和预创建节点
	nodeMap := make(map[string]*DecisionNode)
	for _, nodeCfg := range cfg.Nodes {
		nodeMap[nodeCfg.Name] = &DecisionNode{
			Name: nodeCfg.Name,
		}
	}

	// 构建节点关系和条件函数
	for _, nodeCfg := range cfg.Nodes {
		currentNode := nodeMap[nodeCfg.Name]

		// 解析条件表达式
		rule, exists := cfg.Rules[nodeCfg.Condition]
		if !exists {
			return nil, fmt.Errorf("未定义的规则: %s", nodeCfg.Condition)
		}

		expr, err := govaluate.NewEvaluableExpression(rule.Expr)
		if err != nil {
			return nil, fmt.Errorf("解析表达式失败: %w", err)
		}

		// 创建条件判断函数
		currentNode.Condition = func(order *model.Order) bool {
			parameters := BuildParam(order)
			result, err := expr.Evaluate(parameters)
			if err != nil {
				log.Fatal("条件表达式解析错误: %v", err)
				return false
			}
			return result.(bool)
		}

		// 连接分支节点
		if nodeCfg.TrueBranch != "" {
			if branch, exists := nodeMap[nodeCfg.TrueBranch]; exists {
				currentNode.TrueBranch = branch
			} else if strings.HasPrefix(nodeCfg.TrueBranch, "route:") {
				currentNode.TrueBranch = &DecisionNode{
					RouteDest: strings.TrimPrefix(nodeCfg.TrueBranch, "route:"),
				}
			} else {
				return nil, fmt.Errorf("未定义的True分支节点: %s", nodeCfg.TrueBranch)
			}
		}

		// 同上处理FalseBranch
		// 处理False分支连接
		if nodeCfg.FalseBranch != "" {
			if branch, exists := nodeMap[nodeCfg.FalseBranch]; exists {
				currentNode.FalseBranch = branch
			} else if strings.HasPrefix(nodeCfg.FalseBranch, "route:") {
				currentNode.FalseBranch = &DecisionNode{
					RouteDest: strings.TrimPrefix(nodeCfg.FalseBranch, "route:"),
				}
			} else {
				return nil, fmt.Errorf("未定义的False分支节点: %s", nodeCfg.FalseBranch)
			}
		}
	}

	// 假设第一个节点为根节点
	if len(cfg.Nodes) == 0 {
		return nil, fmt.Errorf("配置文件中没有定义决策树节点")
	}
	rootNode := nodeMap[cfg.Nodes[0].Name]

	return &RuleTree{decisionTree: rootNode}, nil
}

func BuildParam(order *model.Order) map[string]interface{} {
	return map[string]interface{}{
		// 基础字段
		"TotalAmount": order.TotalAmount,
		"UserType":    order.UserType,
		"Location":    order.Location,
		"ItemCount":   len(order.Items),

		// 风控字段 ⭐
		"device_score":       order.DeviceScore,
		"ip_risk_level":      order.IPRiskLevel,
		"geo_mismatch":       order.GeoMismatch,
		"auth_level":         order.AuthLevel,
		"session_age":        order.SessionAge,
		"two_factor_enabled": order.TwoFactorEnabled,

		// 企业验证字段 ⭐
		"licenseValid":       order.LicenseValid,
		"tax_status":         order.TaxStatus,
		"registered_capital": order.RegisteredCapital,

		// 支付相关字段 ⭐
		"payment_method":        order.PaymentMethod,
		"billing_country":       order.BillingCountry,
		"shipping_country":      order.ShippingCountry,
		"is_new_payment_method": order.IsNewPaymentMethod,

		// 物流相关字段 ⭐
		"total_weight":           order.TotalWeight,
		"max_dimension":          order.MaxDimension,
		"contains_fragile_items": order.ContainsFragile,

		// 用户行为字段 ⭐
		"last_login_country":   order.LastLoginCountry,
		"registration_country": order.RegistrationCountry,
		"device_change_count":  order.DeviceChangeCount,

		// 其他衍生字段 ⭐
		"is_enterprise_user": order.UserType == "enterprise",
		"transaction_total": order.TransactionTotal,
		"user_type": order.UserType,
		"risk_score": 100,
		"account_age": 100,
		"enterprise_user" : true,
		"personal_purchases": 5,
	}
}

