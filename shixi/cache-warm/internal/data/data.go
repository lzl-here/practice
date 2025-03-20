package data

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectMySQL() (*gorm.DB, error) {
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

func ConnectRedis() (*redis.Client, error) {
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
