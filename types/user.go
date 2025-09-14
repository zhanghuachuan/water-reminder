package types

import "time"

type User struct {
	ID       string `gorm:"primaryKey" json:"id"`
	Email    string `gorm:"unique;not null" json:"email"`
	Username string `gorm:"not null" json:"username"`
	Password string `gorm:"not null" json:"-"` // 不序列化到JSON
}

type ReminderConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"not null" json:"userId"`
	Enabled     bool      `gorm:"not null;default:true" json:"enabled"`
	StartTime   time.Time `gorm:"not null" json:"startTime"`   // 提醒开始时间(每天)
	EndTime     time.Time `gorm:"not null" json:"endTime"`     // 提醒结束时间(每天)
	Interval    int       `gorm:"not null" json:"interval"`    // 提醒间隔(分钟)
	DailyTarget int       `gorm:"not null" json:"dailyTarget"` // 每日目标(毫升)
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type WaterRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     string    `gorm:"not null" json:"userId"`
	Amount     int       `gorm:"not null" json:"amount"` // 喝水量(毫升)
	RecordTime time.Time `gorm:"not null" json:"recordTime"`
	Action     string    `gorm:"not null" json:"action"` // "drank"或"skipped"
	ReminderID string    `gorm:"not null" json:"reminderId"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
}
