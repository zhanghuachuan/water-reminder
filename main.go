package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/zhanghuachuan/water-reminder/database"
	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/operators"
)

// SchedulerHandler 适配器，使Scheduler兼容http.Handler
type SchedulerHandler struct {
	scheduler *framework.Scheduler
}

func (h *SchedulerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// 从请求中获取 server_name，如果没有则使用默认值 "login"
	serverName := r.URL.Query().Get("server_name")
	if serverName == "" {
		serverName = "login"
	}

	_, err := h.scheduler.Execute(ctx, serverName, framework.ExecuteOptions{
		Request:  r,
		Response: w,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// 1. 加载.env配置
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// 2. 初始化数据库和Redis
	if err := database.InitFromEnv(); err != nil {
		log.Fatal("Failed to initialize databases:", err)
	}

	// 3. 显式初始化算子
	operators.InitializeOperators()

	// 4. 创建调度器并加载配置
	sched := framework.NewScheduler()
	if err := sched.LoadConfig("config/scheduler_config.json"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 5. 启动HTTP服务
	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", &SchedulerHandler{scheduler: sched}); err != nil {
		log.Fatal("Server error:", err)
	}
}
