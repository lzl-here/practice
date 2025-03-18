package shop

import (
	"context"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)


func LoadShop(ctx context.Context, db *gorm.DB, rdb *redis.Client, shopID int) {
	// 尝试获取锁，防止重复
	// 执行预热，每次load 500条
	// 异常退出后写入消息表
}

func CronJob(ch chan int) {
	// 捞取所有异常退出的任务
	// 数据放入队列中，状态改为运行中
}