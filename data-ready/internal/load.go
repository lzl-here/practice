package main

import (
	"context"
	"data-ready/internal/model"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func loadData(ctx context.Context, db *gorm.DB, cache *redis.Client, shopID int) {
	PreheatHandler(ctx, db, cache, shopID)
}

func PreheatHandler(ctx context.Context, db *gorm.DB, cache *redis.Client, shopID int) error {
	// 1. 状态检查
	if status, _ := checkPreheatStatus(ctx, cache, shopID); status != "not_exists" {
		sendAlert(fmt.Sprintf("重复预热请求 shopID:%d", shopID))
		return fmt.Errorf("重复预热请求 shopID:%d", shopID)
	}

	// 2. 设置进行中状态
	if err := setPreheatStatus(ctx, cache, shopID, "processing"); err != nil {
		sendAlert(fmt.Sprintf("状态设置失败 shopID:%d", shopID))
		return fmt.Errorf("状态设置失败 shopID:%d", shopID)
	}

	// 3. 执行预热
	lastID, err := executePreheat(ctx, db, cache, shopID)
	if err != nil {
		handlePreheatFailure(ctx, cache, shopID, lastID, err)
		return fmt.Errorf("预热失败 shopID:%d", shopID)
	}

	// 4. 标记成功
	if err := cache.Set(ctx, fmt.Sprintf(PreheatingKey, shopID), "success", 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("标记成功失败 shopID:%d", shopID)
	}
	return nil
}

func handlePreheatFailure(ctx context.Context, cache *redis.Client, shopID int, userID int, err error) {
	sendAlert(fmt.Sprintf("预热失败 shopID:%d", shopID))
	// 写入缓存进度
	cursorKey := fmt.Sprintf(PreheatingUserCursor, shopID)
	cache.Set(ctx, cursorKey, userID, 24*time.Hour)
}

func executePreheat(ctx context.Context, db *gorm.DB, cache *redis.Client, shopID int) (int, error) {
	cursorKey := fmt.Sprintf(PreheatingUserCursor, shopID)
	lastID, err := cache.Get(ctx, cursorKey).Int()
	if err != nil {
		return 0, err
	}

	for {
		users := make([]model.User, 0)
		err := db.Raw("SELECT * FROM users WHERE shop_id = ? AND id > ? ORDER BY id LIMIT 500", shopID, lastID).Scan(&users).Error

		if err != nil {
			return lastID, err
		}

		if len(users) == 0 {
			break
		}
		lastID = users[len(users)-1].ID

		// 批量写入Redis（需实现batchSetToCache）
		if err := batchSetToCache(ctx, cache, users); err != nil {
			// 记录当前游标
			cache.Set(ctx, cursorKey, lastID, 24*time.Hour)
			return lastID, err
		}

	}

	// 清除游标
	cache.Del(ctx, cursorKey)
	return lastID, nil
}

func batchSetToCache(ctx context.Context, cache *redis.Client, users []model.User) error {
	pipe := cache.Pipeline()

	for _, user := range users {
		key := fmt.Sprintf("user:%d", user.ID)
		// 假设User结构体可以被正确地序列化为JSON
		userJSON, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("序列化用户数据失败: %w", err)
		}

		// 设置24小时过期时间
		pipe.Set(ctx, key, userJSON, 24*time.Hour)
	}

	// 执行管道中的所有命令
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("批量写入Redis失败: %w", err)
	}

	return nil
}

func withRetry(fn func() error, maxRetries int) error {
	retries := 0
	backoff := 1 * time.Second

	for {
		err := fn()
		if err == nil {
			return nil
		}

		retries++
		if retries >= maxRetries {
			return fmt.Errorf("超过最大重试次数")
		}

		time.Sleep(backoff)
		backoff = time.Duration(math.Min(float64(backoff), float64(30*time.Second)))
	}
}

const (
	PreheatingKey        = "preheat:%d:status"      // 店铺ID占位符
	PreheatingUserCursor = "preheat:%d:user_cursor" // 用户游标
)

// 检查预热状态
func checkPreheatStatus(ctx context.Context, cache *redis.Client, shopID int) (string, error) {
	key := fmt.Sprintf(PreheatingKey, shopID)
	status, err := cache.Get(ctx, key).Result()
	if err == redis.Nil {
		return "not_exists", nil
	}
	return status, err
}

// 设置预热状态（使用NX模式防并发）
func setPreheatStatus(ctx context.Context, cache *redis.Client, shopID int, status string) error {
	key := fmt.Sprintf(PreheatingKey, shopID)
	return cache.SetNX(ctx, key, status, 24*time.Hour).Err()
}

func sendAlert(msg string) {
	fmt.Println("发送报警消息：", msg)
}
