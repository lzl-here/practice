package server

import (
	"async-order/internal/consumer"
	"async-order/internal/model"
	"async-order/internal/producer"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

const (
	brokerAddress = "localhost:9092"
	topic         = "order-topic"
)

func AsyncServer(db *gorm.DB, cache *redis.Client) {
	// 初始化Gin
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokerAddress),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()
	r := gin.Default()
	// 添加订单接口
	r.GET("orders/create", func(c *gin.Context) {
		createOrder(c, writer)
	})
	// 启动HTTP服务
	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatal("Gin启动失败:", err)
		}
	}()
	go consumer.StartNormalConsumer(db, cache)
	go consumer.StartDeadConsumer(db, cache)
	select {}
}


// 模拟异步下单
func createOrder(c *gin.Context, writer *kafka.Writer) {
	action := &model.OrderAction{
		AppID:      "APP_" + time.Now().Format("20060102"),
		OrderID:    uuid.New().String(),
		ActionType: "created",
	}

	// 发送到Kafka
	if err := producer.SendKafkaMQ(writer, action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "消息入队失败",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"msg":      "ok",
		"order_id": action.OrderID,
	})
}
