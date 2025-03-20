package shop

import (
	"context"
	"data-ready/internal/model"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func LoadShop(ctx context.Context, db *gorm.DB, rdb *redis.Client, shopID uint64) {
	processID, err := doLoadShop(ctx, db, rdb, shopID)
	if err != nil {
		_ = fmt.Errorf("预热失败 shopID:%d", shopID)
		err = db.Model(&model.Task{}).Updates(&model.Task{
			Status:    model.ErrorStatus(),
			ProcessID: processID,
		}).Error
		if err != nil {
			_ = fmt.Errorf("更新任务状态失败 shopID:%d", shopID)
		}
	}
}

func doLoadShop(ctx context.Context, db *gorm.DB, rdb *redis.Client, shopID uint64) (uint64, error) {
	// 尝试获取锁，防止重复
	if !rdb.SetNX(ctx, "shop_"+strconv.Itoa(int(shopID)), 1, 5*time.Minute).Val() {
		return 0, nil
	}
	// 写入消息表
	var err error
	err = db.Model(&model.Task{}).Save(&model.Task{
		ShopID:    uint64(shopID),
		ProcessID: 0,
		Status:    model.ReadyStatus(),
	}).Error
	if err != nil {
		return 0, err
	}
	raws := make([]*model.User, 1)
	processID := uint64(0)
	// 执行预热，每次load 500条
	for len(raws) > 0 {
		if err = db.Model(&model.User{}).Where("shop_id = ?", shopID).Where("id > ?", processID).Limit(500).Find(raws).Error; err != nil {
			return processID, err
		}
		if err = rdb.MSet(ctx, raws).Err(); err != nil {
			return processID, err
		}
		processID = raws[len(raws)-1].ID
	}
	return 0, nil
}

func CronJob(ch chan int) {
	// 捞取所有异常退出的任务
	// 数据放入队列中，状态改为运行中
}
