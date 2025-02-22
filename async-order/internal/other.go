package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func connectDB() (*gorm.DB, error) {
	hostname := "root"
	password := "376772346Lzl@"
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // 输出到控制台
		logger.Config{
			SlowThreshold:        time.Second, // 慢查询阈值
			LogLevel:             logger.Info, // 日志级别
			ParameterizedQueries: false,
			Colorful:             true, // 彩色打印
		},
	)

	cfg := &gorm.Config{
		Logger: newLogger,
	}
	return gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/order-test", hostname, password)), cfg)
}

func connectRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
}
