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
	UserType    string
	DeviceID    string
	IP          string
	Location    string

	// 新增风控字段 ⭐
	DeviceScore      int    // 设备信用分 0-100
	IPRiskLevel      string // IP风险等级 low/medium/high
	GeoMismatch      bool   // 地理定位不匹配
	AuthLevel        int    // 认证等级 1-5
	SessionAge       int    // 会话存活时间(秒)
	TwoFactorEnabled bool   // 是否启用双因素认证

	// 企业验证字段 ⭐
	LicenseValid      bool    // 营业执照有效性
	TaxStatus         string  // 税务状态 normal/abnormal
	RegisteredCapital float64 // 注册资本

	// 支付相关字段 ⭐
	PaymentMethod      string // 支付方式 crypto/international_card/...
	BillingCountry     string // 账单国家
	ShippingCountry    string // 收货国家
	IsNewPaymentMethod bool   // 是否新支付方式

	// 物流相关字段 ⭐
	TotalWeight     float64 // 总重量(kg)
	MaxDimension    float64 // 最大尺寸(cm)
	ContainsFragile bool    // 是否含易碎品

	// 用户行为字段 ⭐
	LastLoginCountry    string // 最近登录国家
	RegistrationCountry string // 注册国家
	DeviceChangeCount   int    // 设备变更次数

	TransactionTotal int
}

type OrderItem struct {
	SKU         string
	ProductType string // "course", "material", "membership"
	IsVirtual   bool
	Quantity    int
}

func MockOrder() *Order {
	rand.Seed(time.Now().UnixNano())

	order := &Order{
		// 原有字段...

		// 新增字段初始化 ⭐
		DeviceScore:      rand.Intn(101),
		IPRiskLevel:      randomChoice([]string{"low", "medium", "high"}),
		GeoMismatch:      rand.Intn(2) == 1,
		AuthLevel:        rand.Intn(5) + 1,
		SessionAge:       rand.Intn(3600),
		TwoFactorEnabled: rand.Intn(2) == 1,

		LicenseValid:      rand.Intn(2) == 1,
		TaxStatus:         randomChoice([]string{"normal", "abnormal"}),
		RegisteredCapital: rand.Float64()*1000000 + 100000,

		PaymentMethod:      randomChoice([]string{"crypto", "international_card", "alipay", "wechat"}),
		BillingCountry:     randomChoice([]string{"中国", "美国", "日本"}),
		ShippingCountry:    randomChoice([]string{"中国", "美国", "日本"}),
		IsNewPaymentMethod: rand.Intn(2) == 1,

		TotalWeight:     rand.Float64()*50 + 5,
		MaxDimension:    rand.Float64()*200 + 20,
		ContainsFragile: rand.Intn(2) == 1,

		LastLoginCountry:    randomChoice([]string{"中国", "美国", "日本"}),
		RegistrationCountry: randomChoice([]string{"中国", "美国", "日本"}),
		DeviceChangeCount:   rand.Intn(5),
	}

	// 保持原有生成逻辑...
	return order
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
