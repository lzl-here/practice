package main

import (
	"concurrent-select/internal/cache"
	"concurrent-select/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coocood/freecache"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	_ "net/http/pprof"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db = ConnectDB()

func main() {

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	server := gin.New()
	server.Use(gin.Logger())
	// server.Use(gin.Recovery())

	cache := ConnectCache()
	ctx := context.Background()
	handleGet := func(c *gin.Context) {
		id := c.Param("id")
		course, err := getData(ctx, cache, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": course,
		})
	}

	server.GET("/get/:id", handleGet)
	server.Run(":8080")
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
