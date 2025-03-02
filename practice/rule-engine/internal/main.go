package main

import (
	"context"
	"log"
	"rule-engine/internal/model"
	"rule-engine/internal/route"
	"rule-engine/internal/service"
	"runtime"
	"sync"
	"time"
)

func main() {
	// 1. 初始化路由引擎（同原实现）

	router := initRouter()
	// 2. 生成30万测试订单
	orders := GenMockOrders(300_000)

	// parrallelRoute(context.Background(), router, orders)
	normalRoute(context.Background(), router, orders)

}

func initRouter() *route.Router {
	rules := []route.RoutingRule{ /*...*/ }
	engine := route.NewRuleEngine(rules)
	router := &route.Router{
		RuleEngine:     engine,
		ServiceClients: service.NewServiceMap(),
	}
	return router
}

func GenMockOrders(count int) []*model.Order {
	orders := make([]*model.Order, count)
	for i := 0; i < count; i++ {
		orders[i] = model.MockOrder()

	}
	return orders
}

// 新增内存统计函数
func memUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func parrallelRoute(ctx context.Context, router *route.Router, orders []*model.Order) {
	// 3. 并发路由处理
	var wg sync.WaitGroup
	workers := runtime.NumCPU() * 2      // 2倍CPU核心数
	ch := make(chan *model.Order, 10000) // 缓冲队列

	startRoute := time.Now()

	// 启动工作池
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for order := range ch {
				_ = router.RouteOrder(context.Background(), order)
			}
		}(i)
	}

	// 分发任务（分批提交防OOM）
	batchSize := 10_000
	for i := 0; i < len(orders); i += batchSize {
		end := i + batchSize
		if end > len(orders) {
			end = len(orders)
		}
		for _, o := range orders[i:end] {
			ch <- o
		}
	}
	close(ch)
	wg.Wait()
	// 4. 性能报告
	dur := time.Since(startRoute)
	log.Printf("路由完成 | 总耗时: %v | QPS: %.0f/s | 峰值内存: %.2fMB",
		dur,
		float64(len(orders))/dur.Seconds(),
		float64(memUsage())/1024/1024)
}

func normalRoute(ctx context.Context, router *route.Router, orders []*model.Order) {
	// 3. 并发路由处理
	startRoute := time.Now()
	for _, order := range orders {
		_ = router.RouteOrder(context.Background(), order)
	}
	dur := time.Since(startRoute)
	log.Printf("路由完成 | 总耗时: %v | QPS: %.0f/s | 峰值内存: %.2fMB",
		dur,
		float64(len(orders))/dur.Seconds(),
		float64(memUsage())/1024/1024)
}
