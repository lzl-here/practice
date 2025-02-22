package main

import (
	"time"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

// TODO 消费死信队列
// 记录错误记录
func startDeadConsumer(db *gorm.DB) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:       []string{brokerAddress},
		GroupID:       consumerGroupID,
		Topic:         deadTopic,
		MinBytes:      50e3, // 提高最小批量阈值
		MaxBytes:      10e6,
		MaxWait:       500 * time.Millisecond, // 延长批量等待时间
		StartOffset:   kafka.LastOffset,       // 从最新位置开始消费
		QueueCapacity: 200,                    // 预读取队列容量
	})

	defer reader.Close()

}
