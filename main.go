package main

import (
	"log"
	"net/http"

	"github.com/zhanghuachuan/water-reminder/framework/scheduler"
	"github.com/zhanghuachuan/water-reminder/framework/types"
	_ "github.com/zhanghuachuan/water-reminder/operators"
)

func main() {
	// 创建调度器
	sched := scheduler.NewScheduler()

	// 注册算子
	factory := types.GetFactory()
	if op, err := factory.Create("jwt_auth"); err == nil {
		sched.AddOperator(op)
	}
	if op, err := factory.Create("request_validator"); err == nil {
		sched.AddOperator(op)
	}

	// 启动服务
	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", sched)
}
