package operators

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zhanghuachuan/water-reminder/database"
	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/types"
)

type DrinkingRecordOperator struct{}

func (o *DrinkingRecordOperator) Name() string {
	return "drinking-record"
}

func (o *DrinkingRecordOperator) Execute(ctx context.Context, r *http.Request) (context.Context, *framework.OperatorResult) {
	println("开始执行算子:", o.Name())
	println("请求方法:", r.Method)
	println("请求URL:", r.URL.String())

	user := ctx.Value("user").(*types.User)
	println("用户信息:", user.ID, user.Username)

	switch r.Method {
	case http.MethodPost:
		return o.handleCreateRecord(ctx, r, user)
	default:
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Method not allowed", "MethodNotAllowed", http.StatusMethodNotAllowed),
		}
	}
}

func (o *DrinkingRecordOperator) handleCreateRecord(ctx context.Context, r *http.Request, user *types.User) (context.Context, *framework.OperatorResult) {
	var record types.WaterRecord
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Invalid request", "BadRequest", http.StatusBadRequest),
		}
	}

	// 设置用户ID和时间
	record.UserID = user.ID
	if record.RecordTime.IsZero() {
		record.RecordTime = time.Now()
	}

	// 保存到数据库
	if err := database.GetDB().Create(&record).Error; err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Failed to save record", "InternalServerError", http.StatusInternalServerError),
		}
	}

	return ctx, &framework.OperatorResult{
		Data: record,
	}
}
