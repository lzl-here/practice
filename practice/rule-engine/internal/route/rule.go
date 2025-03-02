package route

import (
	"context"
	"fmt"
	"log"
	"rule-engine/internal/model"
	"rule-engine/internal/service"
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
	rules []RoutingRule
	mu    sync.RWMutex
}

func NewRuleEngine(rules []RoutingRule) *RuleEngine {
	return &RuleEngine{rules: rules}
}

func (e *RuleEngine) EvaluateOrder(order *model.Order) *RoutingDecision {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, rule := range e.rules {
		if matched, decision := rule.Evaluate(order); matched {
			return decision
		}
	}

	// 默认路由规则
	return &RoutingDecision{
		OrderID:              order.ID,
		VirtualDestinations:  []string{"DEFAULT_SERVICE"},
		PhysicalDestinations: []string{"DEFAULT_WAREHOUSE"},
	}
}

// ### 路由执行器 ####
type Router struct {
	RuleEngine     *RuleEngine
	ServiceClients map[string]service.ServiceClient
}

func (r *Router) RouteOrder(ctx context.Context, order *model.Order) error {
	decision := r.RuleEngine.EvaluateOrder(order)
	log.Printf("Routing decision: %+v\n", decision)

	// 需要人工审核时中断流程
	if decision.RequiresReview {
		return sendToManualReview(order, decision)
	}

	// 处理虚拟商品
	var wg sync.WaitGroup
	errChan := make(chan error, len(decision.VirtualDestinations)+len(decision.PhysicalDestinations))

	// 处理虚拟服务
	for _, dest := range decision.VirtualDestinations {
		wg.Add(1)
		go func(service string) {
			defer wg.Done()
			if client, ok := r.ServiceClients[service]; ok {
				if err := client.Process(ctx, order); err != nil {
					errChan <- fmt.Errorf("%s error: %w", service, err)
				}
			}
		}(dest)
	}

	// 处理实物物流
	for _, dest := range decision.PhysicalDestinations {
		wg.Add(1)
		go func(warehouse string) {
			defer wg.Done()
			if err := processPhysicalOrder(warehouse, order); err != nil {
				errChan <- fmt.Errorf("%s error: %w", warehouse, err)
			}
		}(dest)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("order processing errors: %v", errors)
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
