package model

type OrderAction struct {
	AppID      string `json:"appId" gorm:"column:app_id"`
	OrderID    string `json:"orderId" gorm:"column:order_id"`
	ActionType string `json:"action_type" gorm:"column:action_type"`
}

func (*OrderAction) TableName() string {
	return "order_action"
}