package test

import (
	"concurrent-select/internal/cache"
	"concurrent-select/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db      *gorm.DB
	cacheDB *cache.Cache
)

func TestMain(m *testing.M) {
	db = ConnectDB()
	cacheDB = ConnectCache()
	m.Run()
}

func TestGetData(t *testing.T) {
	const (
		totalWorkers = 8000 // 总协程数
		batchSize    = 500  // 分批启动数量
		interval     = 5 * time.Millisecond
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	sem := make(chan struct{}, batchSize) // 并发控制信号量

	for i := 0; i < totalWorkers; i++ {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量

		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }()

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// 执行请求（处理错误）
					if _, err := getData(ctx, cacheDB, strconv.Itoa(id)); err != nil {
						t.Logf("worker %d error: %v", id, err)
					}
				}
			}
		}(10)
	}

	wg.Wait()
}

func getData(ctx context.Context, cache *cache.Cache, id string) (*model.Course, error) {
	bytes, err := cache.Get(ctx, "course:"+id, 1000*time.Second, func() (interface{}, error) {
		var err error
		course := new(model.Course)
		if err = db.Model(&model.Course{}).Where("id = ?", id).Find(course).Error; err != nil {
			return nil, err
		}
		return course, nil
	})
	if err != nil {
		return nil, err
	}
	res := new(model.Course)
	json.Unmarshal(bytes, &res)
	return res, nil
}

func ConnectDB() *gorm.DB {
	// 初始化 MySQL 连接
	dsn := "root:376772346Lzl@@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// 配置连接池（可选）
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间
	return db
}

func ConnectCache() *cache.Cache {
	rds := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	localcache := freecache.NewCache(1024 * 1024 * 100) // 100M
	cache := cache.NewCache(rds, localcache)
	return cache
}
