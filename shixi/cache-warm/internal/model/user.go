package model

type User struct {
	ID     uint64    `json:"id" gorm:"column:id"`
	Name   string `json:"name" gorm:"column:name"`
	ShopID int    `json:"shop_id" gorm:"column:shop_id"`
}

func (*User) TableName() string {
	return "user"
}
