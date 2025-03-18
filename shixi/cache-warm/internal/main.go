package main

import (
	"context"
	shop "data-ready/internal/load"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func connectMySQL() (*gorm.DB, error) {
	user := "root"
	password := "376772346Lzl@"
	host := "127.0.0.1"
	port := "3306"
	dbname := "data-ready"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}

func main() {
	db, err := connectMySQL()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	log.Println("Connected to MySQL")

	rdb, err := connectRedis()
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	log.Println("Connected to Redis")

	ctx := context.Background()

	ch := make(chan int, 10000)
	ch <- 1
	consumeNum := 1
	for range consumeNum {
		startConsume(ctx, ch, db, rdb)
	}
	select {}
}

func startConsume(ctx context.Context, ch chan int, db *gorm.DB, rdb *redis.Client) {
	go func() {
		for {
			shopID := <-ch
			shop.LoadShop(ctx, db, rdb, shopID)
		}
	}()
}
