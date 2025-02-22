package consumer

import (
	"async-order/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

// kafka配置

const (
	normalConsumerGroupID = "async-order-consumer-group"
	normalTopic           = "order-topic"

	normalMaxBatchSize   = 50                     // 最大批次大小
	normalFlushInterval  = 100 * time.Millisecond // 批量读取消息等待时间
	normalMaxConcurrency = 60                     // 最大并发批次处理数
)

// 0. 因为需要削峰 需要多协程读取消息然后批量写db

// 1. 在redis中执行setnx ，存在就过滤
// 2. 插入db
// 3. 异常情况：
// 3.1 唯一索引：查db过滤掉已落库的数据, 再次执行批量插入
// 3.2 其他异常：再次投递到kafka，超过一定次数投递到死信队列，记录异常操作

// TODO ack怎么选择？自动？异步？同步？
func StartNormalConsumer(db *gorm.DB, cache *redis.Client) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:       []string{brokerAddress},
		GroupID:       normalConsumerGroupID,
		Topic:         normalTopic,
		MinBytes:      50e3,
		MaxBytes:      500e3,
		StartOffset:   kafka.LastOffset, // 从最新位置开始消费
		QueueCapacity: 200,                // 预读取队列容量
	})
	defer reader.Close()

	abstractConsumer(db, cache, reader, normalMaxConcurrency, normalMaxBatchSize, normalFlushInterval, consumeMsg)
}

// 实际消费的逻辑
func consumeMsg(db *gorm.DB, cache *redis.Client, reader *kafka.Reader, msgs []kafka.Message) {
	if len(msgs) == 0 {
		return
	}
	inserts, err := parseNormalMessages(msgs)
	fmt.Errorf("正在消费消息...")
	if err != nil {
		fmt.Printf("解析消息失败: %v\n", err)
		return
	}
	// 批量写入数据库
	if err = handleOrders(db, inserts); err != nil {
		// panic(err.Error())
		fmt.Errorf("批量插入失败: %v\n", err)
		retryBatch(msgs)
		return
	}

	// 提交偏移量（需保证至少一次语义）
	if err := reader.CommitMessages(context.Background(), msgs...); err != nil {
		fmt.Errorf("提交偏移量失败: %v\n", err)
	}
}

// 过滤掉已经落库的订单
func filterExisting(db *gorm.DB, actions []*model.OrderAction) ([]*model.OrderAction, error) {
	if len(actions) == 0 {
		return nil, nil
	}
	// 构建查询条件
	var params [][]interface{}
	for _, action := range actions {
		params = append(params, []interface{}{action.AppID, action.OrderID, action.ActionType})
	}

	// 执行查询并映射到结构体
	var existing []model.OrderAction
	err := db.Raw(`
	SELECT app_id, order_id, action_type 
	FROM order_action 
	WHERE (app_id, order_id, action_type) IN (?)`, params).
		Scan(&existing).Error

	if err != nil {
		return nil, err
	}

	// 构建存在映射表
	existMap := make(map[string]bool)
	for _, item := range existing {
		key := fmt.Sprintf("%s|%s|%s", item.AppID, item.OrderID, item.ActionType)
		existMap[key] = true
	}
	// 过滤出不存在的记录
	inserts := make([]*model.OrderAction, 0)
	for _, action := range actions {
		key := fmt.Sprintf("%s|%s|%s", action.AppID, action.OrderID, action.ActionType)
		if !existMap[key] {
			inserts = append(inserts, action)
		}
	}

	return inserts, nil
}

// 消费订单
func handleOrders(db *gorm.DB, inserts []*model.OrderAction) error {

	if len(inserts) == 0 {
		return nil
	}

	// 测试唯一索引报错
	// inserts = append(inserts, &model.OrderAction{AppID: "APP_20250222", OrderID: "92bf639d-94e0-4ee8-a44f-230d6990e31f", ActionType: "created"})

	err := db.Create(&inserts).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		if inserts, err = filterExisting(db, inserts); err != nil {
			return err
		}
		if len(inserts) == 0 {
			return nil
		}
		err = db.Create(&inserts).Error
	}
	if err != nil {
		return err
	}
	return nil
}

// 消费失败，重新消费
func retryBatch(batch []kafka.Message) {
	// TODO 投递到kafka中，超过一定次数再投递到死信队列
}

func parseNormalMessages(msgs []kafka.Message) ([]*model.OrderAction, error) {
	var inserts []*model.OrderAction
	for _, msg := range msgs {
		o := new(model.OrderAction)
		if err := json.Unmarshal(msg.Value, o); err != nil {
			return nil, err
		}
		inserts = append(inserts, o)
	}
	return inserts, nil
}
