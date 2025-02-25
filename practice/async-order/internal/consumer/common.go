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
)

type handleFunc func(db *gorm.DB, cache *redis.Client, reader *kafka.Reader, msgs []kafka.Message)

// 抽象出批量读取kafka消息的逻辑
func abstractConsumer(db *gorm.DB, cache *redis.Client, reader *kafka.Reader, maxConcurrency int, maxBatchSize int, flushInternal time.Duration, handler handleFunc) {

	// 把消息从kafka读取到这个chan
	msgChannel := make(chan kafka.Message, maxConcurrency)

	// 消息拉取协程组，这里应该多协程拉取加快读取速度
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
		timeout := time.After(flushInternal)

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

		go func(msgs []kafka.Message) {
			handler(db, cache, reader, msgs)
		}(msgs)
		// time.Sleep(50 * time.Millisecond)
	}
}
