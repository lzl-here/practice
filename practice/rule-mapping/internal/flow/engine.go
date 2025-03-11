package flow

import (
	"fmt"
	"io/ioutil"

	"github.com/Knetic/govaluate"
	"gopkg.in/yaml.v2"
)

// Node 定义流程节点
type Node struct {
	Name        string       `yaml:"name"`
	Action      string       `yaml:"action"`
	Transitions []Transition `yaml:"transitions"`
}

// Transition 定义节点之间的跳转条件
type Transition struct {
	Condition string `yaml:"condition"` // govalue表达式
	Next      string `yaml:"next"`
}

// Workflow 定义流程
type Workflow struct {
	Name  string `yaml:"name"`
	Nodes []Node `yaml:"nodes"`
}

// Engine 定义流程引擎
type Engine struct {
	Workflow   Workflow
	NodeMap    map[string]Node
	ActionFunc map[string]func(*FlowContext) error
}

// NewEngine 创建新的流程引擎
func NewEngine(configPath string) (*Engine, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("解析YAML配置失败: %w", err)
	}

	nodeMap := make(map[string]Node)
	for _, node := range workflow.Nodes {
		nodeMap[node.Name] = node
	}

	return &Engine{
		Workflow:   workflow,
		NodeMap:    nodeMap,
		ActionFunc: make(map[string]func(*FlowContext) error),
	}, nil
}

// RegisterAction 注册节点对应的业务逻辑
func (e *Engine) RegisterAction(action string, fn func(*FlowContext) error) {
	e.ActionFunc[action] = fn
}

// Execute 执行流程
func (e *Engine) Execute(startNode string, contx *FlowContext) error {
	currentNode := e.NodeMap[startNode]

	for {
		actionFn, exists := e.ActionFunc[currentNode.Action]
		// 执行节点业务逻辑
		if exists {
			if err := actionFn(contx); err != nil {
				return fmt.Errorf("节点 %s 执行失败: %w", currentNode.Name, err)
			}
		}

		// 根据结果跳转到下一个节点
		var nextNodeName string
		for _, transition := range currentNode.Transitions {
			expr, err := govaluate.NewEvaluableExpression(transition.Condition)
			if err != nil {
				return err
			}
			if result, err := expr.Evaluate(contx.ToMap()); err == nil {
				if result.(bool) {
					nextNodeName = transition.Next
					break
				}
			}
		}

		if nextNodeName == "" {
			break // 流程结束
		}

		nextNode, exists := e.NodeMap[nextNodeName]
		if !exists {
			return fmt.Errorf("未定义的节点: %s", nextNodeName)
		}

		currentNode = nextNode
	}

	return nil
}
