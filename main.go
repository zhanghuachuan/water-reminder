package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/zhanghuachuan/water-reminder/database"
	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/operators"
	"github.com/zhanghuachuan/water-reminder/types"
)

// SchedulerHandler 适配器，使Scheduler兼容http.Handler
type SchedulerHandler struct {
	scheduler *framework.Scheduler
}

// ResponseHandler 统一的HTTP响应处理器
type ResponseHandler struct {
	StatusCode int
	Data       interface{}
	Error      error
}

// handleResponse 统一处理HTTP响应
func handleResponse(w http.ResponseWriter, results []framework.ExecutionResult) {
	w.Header().Set("Content-Type", "application/json")

	// 检查是否有错误
	for _, result := range results {
		if result.Error != nil {
			// 检查是否是ApiError类型
			if apiErr, ok := result.Error.(*types.ApiError); ok {
				w.WriteHeader(apiErr.StatusCode)
				json.NewEncoder(w).Encode(types.NewErrorResponse(apiErr.Message, apiErr.ErrorMsg))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(types.NewErrorResponse("执行失败", result.Error.Error()))
			}
			return
		}
	}

	// 获取最后一个算子的结果数据
	if len(results) > 0 {
		lastResult := results[len(results)-1]
		// 这里可以根据需要从上下文中提取数据或使用其他方式获取最终结果
		if lastResult.Error != nil {
			// 检查是否是ApiError类型
			if apiErr, ok := lastResult.Error.(*types.ApiError); ok {
				w.WriteHeader(apiErr.StatusCode)
				json.NewEncoder(w).Encode(types.NewErrorResponse(apiErr.Message, apiErr.ErrorMsg))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(types.NewErrorResponse("执行失败", lastResult.Error.Error()))
			}
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(types.NewSuccessResponse("操作成功", lastResult.Data))
		}
		return
	}

	// 默认成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(types.NewSuccessResponse("操作成功", nil))
}

func (h *SchedulerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// 从URL路径中获取 server_name
	// 路径格式例如: /api/login, /api/register 等
	serverName := r.URL.Path // 默认值

	// 从路径中提取server_name，例如从 "/login" 中提取
	if serverName == "" {
		serverName = "login"
	}

	results, err := h.scheduler.Execute(ctx, serverName, framework.ExecuteOptions{
		Request: r,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.NewErrorResponse("调度失败", err.Error()))
		return
	}

	// 统一处理HTTP响应
	handleResponse(w, results)
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
