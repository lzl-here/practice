package producer

import (
	"context"
	"fmt"
	"async-order/internal/model"

	"encoding/json"

	"github.com/segmentio/kafka-go"
)

// 生产者
// 1. 写入数据到缓存，写入失败返回下单失败
// 2. 发送kafka消息，发送失败返回失败
// 两者都可以加上最大重试次数，超过次数告警 （优化：自动退避重试）

func SendKafkaMQ(writer *kafka.Writer, oa *model.OrderAction) error {

	bytes, err := json.Marshal(oa)
	if err != nil {
		fmt.Printf("消息序列化失败: %v\n", err)
		return err
	}
	msg := kafka.Message{
		Key:   []byte("Key"),
		Value: bytes,
	}

	err = writer.WriteMessages(context.Background(), msg)
	if err != nil {
		fmt.Printf("生产者错误: %v\n", err)
		return err
	}

	fmt.Printf("已发送消息: %s\n", msg.Value)
	return err
}
