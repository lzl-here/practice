package consumer

import (
	"async-order/internal/model"
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

const (
	deadConsumerGroupID = "async-order-dead-consumer-group"
	deadTopic           = "dead-order-topic"

	deadMaxBatchSize   = 10
	deadMaxConcurrency = 5
	deadFlushInterval  = 100 * time.Millisecond
)

// TODO 消费死信队列
// 记录错误记录
func StartDeadConsumer(db *gorm.DB, cache *redis.Client) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:       []string{brokerAddress},
		GroupID:       deadConsumerGroupID,
		Topic:         deadTopic,
		MinBytes:      50e3, // 提高最小批量阈值
		MaxBytes:      10e6,
		MaxWait:       500 * time.Millisecond, // 延长批量等待时间
		StartOffset:   kafka.LastOffset,       // 从最新位置开始消费
		QueueCapacity: 200,                    // 预读取队列容量
	})
	defer reader.Close()

	abstractConsumer(db, cache, reader, deadMaxConcurrency, deadMaxBatchSize, deadFlushInterval, consumeDeadMsg)
}

func parseDeadMessage(msgs []kafka.Message) ([]*model.OrderRecord, error) {

	var inserts []*model.OrderRecord
	for _, msg := range msgs {
		o := new(model.OrderRecord)
		if err := json.Unmarshal(msg.Value, o); err != nil {
			return nil, err
		}
		inserts = append(inserts, o)
	}

	// 示例：假设消息是JSON格式
	return inserts, nil
}

func consumeDeadMsg(db *gorm.DB, cache *redis.Client, reader *kafka.Reader, msgs []kafka.Message) {

}
