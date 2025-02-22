package main

import (
	"async-order/internal/model"
	"log"
	"net/http"
	"time"

	consumer "async-order/internal/consumer"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

const (
	brokerAddress = "localhost:9092"
	topic         = "order-topic"
	deadTopic     = "dead-order-topic"
)

// 为什么需要做异步下单？
// 问题在于db的负载过高，需要减少db的压力：
// db承载的了多少qps的请求？ 其中读请求qps多少？ 写请求qps多少？

// 读方面: 原先就有缓存，所以这里不怎么需要优化读方面

// 写方面: 现在db的压力主要集中在写上面
// 通过kafka来做数据的写聚合，通过批量写入的方式降低写入db的频率
// 缺点是会带来一定的延迟，但是这几十ms对整体链路和用户体验影响不大

func main() {
	db, err := connectDB()
	cache := connectRedis()
	if err != nil {
		panic(err)
	}
	// 初始化Gin
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokerAddress),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()
	startServer(writer)
	go consumer.StartNormalConsumer(db, cache)
	go consumer.StartDeadConsumer(db, cache)
	select {}
}

// 模拟下单的接口x
func createOrder(c *gin.Context, writer *kafka.Writer) {
	action := &model.OrderAction{
		AppID:      "APP_" + time.Now().Format("20060102"),
		OrderID:    uuid.New().String(),
		ActionType: "created",
	}

	// 发送到Kafka
	if err := sendKafkaMQ(writer, action); err != nil {
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

func startServer(writer *kafka.Writer) {
	router := gin.Default()
	// 添加订单接口
	orderGroup := router.Group("/orders")
	{
		orderGroup.GET("create", func(c *gin.Context) {
			createOrder(c, writer)
		})
	}

	// 启动HTTP服务
	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatal("Gin启动失败:", err)
		}
	}()

}
