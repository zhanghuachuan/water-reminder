package models

import "gorm.io/gorm"

type WaterRecord struct {
	gorm.Model
	UserID    uint    `gorm:"not null"`
	Amount    float32 `gorm:"not null"` // 毫升
	Time      int64   `gorm:"not null"` // 时间戳
	DrinkType string  // 饮品类型
}
