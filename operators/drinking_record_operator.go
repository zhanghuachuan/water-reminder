package operators

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zhanghuachuan/water-reminder/database"
	"github.com/zhanghuachuan/water-reminder/types"
)

type DrinkingRecordOperator struct{}

func (o *DrinkingRecordOperator) Name() string {
	return "drinking-record"
}

func (o *DrinkingRecordOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	println("开始执行算子:", o.Name())
	println("请求方法:", r.Method)
	println("请求URL:", r.URL.String())

	user := ctx.Value("user").(*types.User)
	println("用户信息:", user.ID, user.Username)

	var resultCtx context.Context
	var err error

	switch r.Method {
	case http.MethodPost:
		resultCtx, err = o.handleCreateRecord(ctx, w, r, user)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		resultCtx, err = ctx, nil
	}

	if err != nil {
		println("算子执行失败:", err.Error())
	} else {
		println("算子执行成功")
	}

	return resultCtx, err
}

func (o *DrinkingRecordOperator) handleCreateRecord(ctx context.Context, w http.ResponseWriter, r *http.Request, user *types.User) (context.Context, error) {
	var record types.WaterRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return ctx, err
	}

	// 设置用户ID和时间
	record.UserID = user.ID
	if record.RecordTime.IsZero() {
		record.RecordTime = time.Now()
	}

	// 保存到数据库
	if err := database.GetDB().Create(&record).Error; err != nil {
		http.Error(w, "Failed to save record", http.StatusInternalServerError)
		return ctx, err
	}

	w.WriteHeader(http.StatusCreated)
	return ctx, nil
}
