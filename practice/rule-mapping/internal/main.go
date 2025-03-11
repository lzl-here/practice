package main

import (
	"fmt"
	"log"
	"rule-mapping/internal/flow"
	"rule-mapping/internal/mapping"
)

func main() {
	engine := initFlowEngine()
	contx := mapping.BuildCtx()
	// 执行流程
	if err := engine.Execute("订单处理流程", contx); err != nil {
		log.Fatalf("流程执行失败: %v", err)
	}
}

func initFlowEngine() *flow.Engine {
	// 初始化流程引擎
	engine, err := flow.NewEngine("workflow.yaml")
	if err != nil {
		log.Fatalf("初始化引擎失败: %v", err)
	}

	// 注册节点对应的业务逻辑
	engine.RegisterAction("validateOrder", func(contx *flow.FlowContext) error {
		fmt.Println("执行: 验证订单")
		return nil
	})

	engine.RegisterAction("processPayment", func(contx *flow.FlowContext) error {
		fmt.Println("执行: 处理支付")
		return nil
	})

	engine.RegisterAction("generateShippingOrder", func(contx *flow.FlowContext) error {
		fmt.Println("执行: 生成运单")
		return nil
	})

	engine.RegisterAction("failOrder", func(contx *flow.FlowContext) error {
		fmt.Println("执行: 终止流程")
		return nil
	})

	engine.RegisterAction("completeOrder", func(contx *flow.FlowContext) error {
		fmt.Println("执行: 结束流程")
		return nil
	})
	return engine
}
