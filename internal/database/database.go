package database

import (
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/zhanghuachuan/water-reminder/internal/models"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	MySQL struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
}

var DB *gorm.DB
var RDB *redis.Client

func InitDB() error {
	// 读取配置文件
	data, err := os.ReadFile("config/database.yaml")
	if err != nil {
		return err
	}

	var config DatabaseConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// 初始化MySQL
	if err := InitMySQL(config.MySQL.DSN); err != nil {
		return err
	}

	// 初始化Redis
	InitRedis(config.Redis.Addr, config.Redis.Password, config.Redis.DB)

	return nil
}

func InitMySQL(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移模型
	err = DB.AutoMigrate(&models.User{}, &models.WaterRecord{})
	if err != nil {
		log.Printf("AutoMigrate error: %v", err)
		return err
	}

	return nil
}

func InitRedis(addr, password string, db int) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}