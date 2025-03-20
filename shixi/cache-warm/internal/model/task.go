package model

type Task struct {
	ID        uint64 `gorm:"column:id"`
	ShopID    uint64 `gorm:"column:shop_id"`
	ProcessID uint64 `gorm:"column:process_id"`
	Status    int    `gorm:"column:status"`
}

func (*Task) TableName() string {
	return "task"
}

func ReadyStatus() int {
	return 0
}

func RunningStatus() int {
	return 1
}

func ErrorStatus() int {
	return 2
}
