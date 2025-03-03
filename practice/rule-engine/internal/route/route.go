package route

import (
	"context"
	"fmt"
	"log"
	"rule-engine/internal/model"
	"rule-engine/internal/rule"
	"sync"
)

// 路由决策结果
type RoutingDecision struct {
	OrderID              string
	VirtualDestinations  []string // 虚拟商品处理目标系统
	PhysicalDestinations []string // 实物商品处理目标系统
	RequiresReview       bool     // 需要人工审核
	SplitOrders          []string // 拆单后的子订单ID
}

func NewRoutingDecision(order *model.Order, routeDest string) (*RoutingDecision, error) {
	if routeDest == "" {
		return nil, fmt.Errorf("决策规则为空")
	}

	// TODO 由规则构建路由节点
	return &RoutingDecision{
		OrderID:              order.ID,
		VirtualDestinations:  []string{routeDest},
		PhysicalDestinations: []string{routeDest},
		RequiresReview:       false,
		SplitOrders:          []string{},
	}, nil
}

// 规则引擎接口
type RoutingRule interface {
	Evaluate(order *model.Order) (bool, *RoutingDecision)
}

//### 核心规则实现 ####

// 混合订单拆分规则
type MixedOrderRule struct{}

func (r *MixedOrderRule) Evaluate(order *model.Order) (bool, *RoutingDecision) {
	hasVirtual, hasPhysical := false, false
	for _, item := range order.Items {
		if item.IsVirtual {
			hasVirtual = true
		} else {
			hasPhysical = true
		}
	}

	if hasVirtual && hasPhysical {
		decision := &RoutingDecision{
			OrderID:              order.ID,
			VirtualDestinations:  []string{"COURSE_SERVICE", "MEMBERSHIP_SERVICE"},
			PhysicalDestinations: []string{"CENTRAL_WAREHOUSE"},
			SplitOrders:          generateSubOrderIDs(order.ID),
		}
		return true, decision
	}
	return false, nil
}

// 企业订单规则
type EnterpriseRule struct {
	EnterpriseThreshold float64
}

func (r *EnterpriseRule) Evaluate(order *model.Order) (bool, *RoutingDecision) {
	if order.UserType == "enterprise" || order.TotalAmount > r.EnterpriseThreshold {
		return true, &RoutingDecision{
			OrderID:              order.ID,
			VirtualDestinations:  []string{"ENTERPRISE_SERVICE"},
			PhysicalDestinations: []string{"ENTERPRISE_WAREHOUSE"},
			RequiresReview:       order.TotalAmount > 50000,
		}
	}
	return false, nil
}

// 风控规则
type RiskControlRule struct {
	HighRiskLocations map[string]bool
}

func (r *RiskControlRule) Evaluate(order *model.Order) (bool, *RoutingDecision) {
	// 地理位置风险检查
	if _, ok := r.HighRiskLocations[order.Location]; ok {
		return true, &RoutingDecision{
			OrderID:        order.ID,
			RequiresReview: true,
		}
	}

	// 设备指纹逻辑
	if isSuspiciousDevice(order.DeviceID) {
		return true, &RoutingDecision{
			OrderID:        order.ID,
			RequiresReview: true,
		}
	}

	return false, nil
}

// ### 规则引擎 ####
type RuleEngine struct {
	mu   *sync.RWMutex
	tree *rule.RuleTree
}

func NewRuleEngine(tree *rule.RuleTree) *RuleEngine {
	return &RuleEngine{tree: tree, mu: &sync.RWMutex{}}
}

func (e *RuleEngine) EvaluateOrder(order *model.Order) (*RoutingDecision, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	dest, err := e.tree.Match(order)
	if err != nil {
		return nil, err
	}
	return NewRoutingDecision(order, dest)
}

// ### 路由执行器 ####
type Router struct {
	RuleEngine *RuleEngine
}

func (r *Router) RouteOrder(ctx context.Context, order *model.Order) error {
	decision, err := r.RuleEngine.EvaluateOrder(order)
	if err != nil {
		return err
	}
	//TODO 订单路由
	// 根据规则引擎匹配的决策进行后续处理，比如 课程服务需要写入权益、实物商品发物流、大金额订单需要风控系统介入、免费试听订单需要走其他渠道....

	// 需要人工审核时中断流程
	for _, d := range decision.PhysicalDestinations {
		fmt.Println(fmt.Sprintln("订单路由, 实物商品: %s", d))
	}

	for _, v := range decision.VirtualDestinations {
		fmt.Println(fmt.Sprintf("订单路由, 虚拟商品: %s", v))
	}
	return nil
}

// ### 辅助函数 ####
func generateSubOrderIDs(parentID string) []string {
	return []string{
		parentID + "-VIRTUAL",
		parentID + "-PHYSICAL",
	}
}

func isSuspiciousDevice(deviceID string) bool {
	// 实现设备风险检查逻辑
	return false
}

func processPhysicalOrder(warehouse string, order *model.Order) error {
	log.Printf("Processing physical order at %s\n", warehouse)
	// ... 具体仓库处理逻辑
	return nil
}

func sendToManualReview(order *model.Order, decision *RoutingDecision) error {
	log.Printf("Sending order %s to manual review", order.ID)
	// ... 通知审核系统逻辑
	return nil
}
