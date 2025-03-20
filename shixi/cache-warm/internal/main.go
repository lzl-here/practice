package main

import (
	"context"
	"data-ready/internal/data"
	shop "data-ready/internal/load"
	"log"

	"github.com/go-redis/redis/v8"

	"gorm.io/gorm"
)

func main() {
	db, err := data.ConnectMySQL()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	rdb, err := data.ConnectRedis()
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}

	ctx := context.Background()

	ch := make(chan uint64, 10000)
	ch <- 1
	consumeNum := 1
	for range consumeNum {
		go func() {
			startConsume(ctx, ch, db, rdb)
		}()
	}
	select {}
}

func startConsume(ctx context.Context, ch chan uint64, db *gorm.DB, rdb *redis.Client) {
	for {
		shopID := <-ch
		shop.LoadShop(ctx, db, rdb, shopID)
	}
}
