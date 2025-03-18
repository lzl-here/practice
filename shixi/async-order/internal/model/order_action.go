package model

type OrderAction struct {
	AppID      string `json:"app_id" gorm:"column:app_id"`
	OrderID    string `json:"order_id" gorm:"column:order_id"`
	ActionType string `json:"action_type" gorm:"column:action_type"`
}

func (*OrderAction) TableName() string {
	return "order_action"
}