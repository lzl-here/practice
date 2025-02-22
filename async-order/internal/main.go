package main

import (
	data "async-order/internal/data"
	server "async-order/internal/server"
)

func main() {

	db, err := data.ConnectDB()
	cache := data.ConnectRedis()
	if err != nil {
		panic(err)
	}

	server.AsyncServer(db, cache) // 异步下单
	// server.SyncServer(db, cache) // 同步下单
}
