package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

const (
	brokerAddress = "localhost:9092"

	maxBatchSize   = 10                     // 最大批次大小
	flushInterval  = 100 * time.Millisecond // 批量读取消息等待时间
	maxConcurrency = 5                      // 最大并发批次处理数

)

type handleFunc func(db *gorm.DB, cache *redis.Client, reader *kafka.Reader, batchPool chan struct{}, msgChannel chan kafka.Message, msgs []kafka.Message)

// 抽象出批量读取kafka消息的逻辑
func abstractConsumer(db *gorm.DB, cache *redis.Client, reader *kafka.Reader, batchPool chan struct{}, msgChannel chan kafka.Message, handler handleFunc) {

	// 消息拉取协程组
	go func() {
		for {
			msg, err := reader.FetchMessage(context.Background())
			if err != nil {
				fmt.Printf("拉取消息失败: %v\n", err)
				time.Sleep(1 * time.Second) // 错误退避
				continue
			}
			msgChannel <- msg
		}
	}()

	// 批量处理协程
	for {
		msgs := make([]kafka.Message, 0, maxBatchSize)
		timeout := time.After(flushInterval)

		// 聚合批次，拉取到10条消息 / 超时没读取到消息 执行批量消费
	AggregateLoop:
		for {
			select {
			case msg := <-msgChannel:
				msgs = append(msgs, msg)
				if len(msgs) >= maxBatchSize {
					break AggregateLoop
				}
			case <-timeout:
				break AggregateLoop
			}
		}

		if len(msgs) == 0 {
			continue
		}

		batchPool <- struct{}{} // 获取处理槽位
		go func(batch []kafka.Message) {
			handler(db, cache, reader, batchPool, msgChannel, batch)
		}(msgs)
	}
}
