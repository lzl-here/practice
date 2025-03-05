package rule

import (
	"fmt"
	"rule-mapping/internal/model"
	"strings"

	"github.com/Knetic/govaluate"
)

// 决策树节点结构
type DecisionNode struct {
	Name        string
	Condition   func(*model.Order) bool // 判断条件函数
	TrueBranch  *DecisionNode           // 条件为真时的分支
	FalseBranch *DecisionNode           // 条件为假时的分支
	RouteDest   string                  // 叶子节点的路由结果
}

// 路由引擎结构
type RuleTree struct {
	decisionTree *DecisionNode
}

const (
	MaxDepth = 200
)

type TreeConfig struct {
	Nodes []struct {
		Name        string `yaml:"name"`
		Condition   string `yaml:"condition"`
		TrueBranch  string `yaml:"true_branch"`
		FalseBranch string `yaml:"false_branch"`
	} `yaml:"decision_tree"`

	Rules map[string]struct {
		Type string `yaml:"type"` // script/function
		Expr string `yaml:"expr"`
	} `yaml:"rules"`
}

func buildSampleDecisionTree(cfg *TreeConfig) *RuleTree {
	nodeMap := make(map[string]*DecisionNode)

	// 第一阶段：预创建所有节点
	for _, nodeCfg := range cfg.Nodes {
		nodeMap[nodeCfg.Name] = &DecisionNode{
			Name: nodeCfg.Name,
		}
	}

	// 第二阶段：构建节点关系
	for _, nodeCfg := range cfg.Nodes {
		currentNode := nodeMap[nodeCfg.Name]

		// 绑定条件函数
		if rule, exists := cfg.Rules[nodeCfg.Condition]; exists {
			// 使用govaluate动态解析表达式（需导入）
			expr, _ := govaluate.NewEvaluableExpression(rule.Expr)
			currentNode.Condition = func(order *model.Order) bool {
				parameters := map[string]interface{}{
					"TotalAmount": order.TotalAmount,
					"UserType":    order.UserType,
					"Location":    order.Location,
					"ItemCount":   len(order.Items),
				}
				result, _ := expr.Evaluate(parameters)
				return result.(bool)
			}
		}

		// 连接True分支
		if nodeCfg.TrueBranch != "" {
			if branch, exists := nodeMap[nodeCfg.TrueBranch]; exists {
				currentNode.TrueBranch = branch
			} else if strings.HasPrefix(nodeCfg.TrueBranch, "route:") {
				currentNode.TrueBranch = &DecisionNode{
					RouteDest: strings.TrimPrefix(nodeCfg.TrueBranch, "route:"),
				}
			}
		}

		// 连接False分支
		if nodeCfg.FalseBranch != "" {
			if branch, exists := nodeMap[nodeCfg.FalseBranch]; exists {
				currentNode.FalseBranch = branch
			} else if strings.HasPrefix(nodeCfg.FalseBranch, "route:") {
				currentNode.FalseBranch = &DecisionNode{
					RouteDest: strings.TrimPrefix(nodeCfg.FalseBranch, "route:"),
				}
			}
		}
	}

	// 假设第一个节点为根节点
	return &RuleTree{
		decisionTree: nodeMap[cfg.Nodes[0].Name],
	}
}

// 示例条件判断方法
func isHighRiskOrder(order *model.Order) bool {
	return order.TotalAmount > 100000 ||
		(order.UserType == "enterprise" && order.TotalAmount > 50000)
}

func isEnterpriseOrder(order *model.Order) bool {
	return order.UserType == "enterprise" &&
		len(order.Items) > 0 &&
		order.Items[0].Quantity >= 50
}

func hasMixedProducts(order *model.Order) bool {
	hasVirtual, hasPhysical := false, false
	for _, item := range order.Items {
		if item.IsVirtual {
			hasVirtual = true
		} else {
			hasPhysical = true
		}
	}
	return hasVirtual && hasPhysical
}

func isEmergencyArea(order *model.Order) bool {
	return order.Location == "中国-台湾" ||
		order.Location == "中国-南海诸岛"
}

func (e *RuleTree) Match(order *model.Order) (string, error) {
	currentNode := e.decisionTree
	path := make([]string, 0)
	depth := 0

	for currentNode.RouteDest == "" {
		if depth >= MaxDepth {
			return "", fmt.Errorf("决策树深度超过%d层，检测到潜在循环", MaxDepth)
		}
		depth++
		for currentNode.RouteDest == "" {
			path = append(path, currentNode.Name)
			if currentNode.Condition(order) {
				currentNode = currentNode.TrueBranch
			} else {
				currentNode = currentNode.FalseBranch
			}

		}
	}
	return currentNode.RouteDest, nil
}
