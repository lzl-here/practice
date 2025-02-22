package main

import (
	data "async-order/internal/data"
	server "async-order/internal/server"
)

// 为什么需要做异步下单？
// 问题在于db的负载过高，需要减少db的压力：
// db承载的了多少qps的请求？ 其中读请求qps多少？ 写请求qps多少？

// 读方面: 原先就有缓存，所以这里不怎么需要优化读方面

// 写方面: 现在db的压力主要集中在写上面
// 通过kafka来做数据的写聚合，通过批量写入的方式降低写入db的频率
// 缺点是会带来一定的延迟，但是这几十ms对整体链路和用户体验影响不大

func main() {
	db, err := data.ConnectDB()
	cache := data.ConnectRedis()
	if err != nil {
		panic(err)
	}

	server.AsyncServer(db, cache) // 异步下单
	// server.SyncServer(db, cache)  // 同步下单
}
