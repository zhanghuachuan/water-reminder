package services

import (
	"time"

	"github.com/zhanghuachuan/water-reminder/internal/database"
	"github.com/zhanghuachuan/water-reminder/internal/models"
)

func CreateWaterRecord(userID uint, amount float32, drinkType string) (*models.WaterRecord, error) {
	record := models.WaterRecord{
		UserID:    userID,
		Amount:    amount,
		Time:      time.Now().Unix(),
		DrinkType: drinkType,
	}

	result := database.DB.Create(&record)
	if result.Error != nil {
		return nil, result.Error
	}

	return &record, nil
}

func GetUserRecords(userID uint) ([]models.WaterRecord, error) {
	var records []models.WaterRecord
	result := database.DB.Where("user_id = ?", userID).Find(&records)
	return records, result.Error
}

func GetTodayIntake(userID uint) (float32, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Unix()

	var total float32
	err := database.DB.Model(&models.WaterRecord{}).
		Where("user_id = ? AND time BETWEEN ? AND ?", userID, start, end).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error

	return total, err
}