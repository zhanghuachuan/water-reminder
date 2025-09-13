package main

import (
	"trpc.group/trpc-go/trpc-go/log"

	"github.com/zhanghuachuan/water-reminder/internal/database"
	"github.com/zhanghuachuan/water-reminder/internal/trpcservices"
	"trpc.group/trpc-go/trpc-go/server"
)

func main() {
	// 初始化数据库连接
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 创建TRPC服务
	s := server.New()

	// 注册服务
	trpcservices.RegisterAuthService(s, &trpcservices.AuthService{})
	trpcservices.RegisterWaterRecordService(s, &trpcservices.WaterRecordService{})

	// 启动服务
	if err := s.Serve(); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
