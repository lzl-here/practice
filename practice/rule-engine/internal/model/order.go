package model

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// 订单数据结构
type Order struct {
	ID          string
	UserID      int64
	Items       []OrderItem
	TotalAmount float64
	UserType    string // "personal", "enterprise"
	DeviceID    string
	IP          string
	Location    string // 格式："国家-省份-城市"
}

type OrderItem struct {
	SKU         string
	ProductType string // "course", "material", "membership"
	IsVirtual   bool
	Quantity    int
}

func MockOrder() *Order {
	rand.Seed(time.Now().UnixNano()) // 初始化随机种子

	return &Order{
		ID:          generateRandomString(10),
		UserID:      rand.Int63n(1000000),
		TotalAmount: rand.Float64() * 1000,
		UserType:    randomChoice([]string{"personal", "enterprise"}),
		DeviceID:    "DEV-" + generateRandomString(6),
		IP:          generateRandomIP(),
		Location:    randomLocation(),
		Items:       generateRandomItems(),
	}
}

// 辅助函数定义
func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func randomChoice(options []string) string {
	return options[rand.Intn(len(options))]
}

func generateRandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func randomLocation() string {
	countries := []string{"中国", "美国", "日本"}
	provinces := map[string][]string{
		"中国": {"浙江", "广东", "江苏"},
		"美国": {"加州", "德州", "纽约"},
		"日本": {"东京", "大阪", "北海道"},
	}
	country := randomChoice(countries)
	return fmt.Sprintf("%s-%s-%s",
		country,
		randomChoice(provinces[country]),
		generateRandomString(2)+"市")
}

func generateRandomItems() []OrderItem {
	count := rand.Intn(3) + 1 // 生成1-3个item
	items := make([]OrderItem, count)

	for i := range items {
		items[i] = OrderItem{
			SKU:         strings.ToUpper(generateRandomString(6)),
			ProductType: randomChoice([]string{"course", "material", "membership"}),
			IsVirtual:   rand.Intn(2) == 1,
			Quantity:    rand.Intn(5) + 1,
		}
	}
	return items
}
