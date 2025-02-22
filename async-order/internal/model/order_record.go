package model

// 订单流水表，记录一些信息，比如错误
type OrderRecord struct {
	AppID        string `json:"appId" gorm:"column:app_id"`
	OrderID      string `json:"orderId" gorm:"column:order_id"`
	FailedReason string `json:"failedReason" gorm:"column:failed_reason"`
}

func (*OrderRecord) TableName() string {
	return "order_record"
}
