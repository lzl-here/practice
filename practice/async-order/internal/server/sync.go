package server

import (
	"async-order/internal/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SyncServer(db *gorm.DB, cache *redis.Client) {
	// 1. 创建Gin实例
	r := gin.Default()

	r.GET("/orders/create", func(c *gin.Context) {
		action := &model.OrderAction{
			AppID:      "APP_" + time.Now().Format("20060102"),
			OrderID:    uuid.New().String(),
			ActionType: "created",
		}

		if err := db.Create(action).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

	})

	// 3. 启动服务（默认8080端口）
	if err := r.Run(); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
