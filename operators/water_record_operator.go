package operators

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/types"
	"github.com/zhanghuachuan/water-reminder/utils"
)

func init() {
	framework.RegisterOperator("water_record", &WaterRecordOperator{})
}

type WaterRecordOperator struct{}

func (o *WaterRecordOperator) Name() string {
	return "water_record"
}

type WaterRecordRequest struct {
	Amount    float64   `json:"amount"`    // 饮水量（毫升）
	Time      time.Time `json:"time"`      // 饮水时间
	DrinkType string    `json:"drinkType"` // 饮品类型（水/茶/咖啡等）
}

type WaterRecordResponse struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Time      time.Time `json:"time"`
	DrinkType string    `json:"drinkType"`
}

func (o *WaterRecordOperator) Execute(ctx context.Context, r *http.Request) (context.Context, *framework.OperatorResult) {
	// 获取当前用户
	user, ok := ctx.Value("user").(*utils.User)
	if !ok || user == nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Unauthorized", "Unauthorized", http.StatusUnauthorized),
		}
	}

	switch r.Method {
	case http.MethodPost:
		return o.handleCreateRecord(ctx, r, user)
	case http.MethodGet:
		return o.handleGetRecords(ctx, r, user)
	default:
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Method not allowed", "MethodNotAllowed", http.StatusMethodNotAllowed),
		}
	}
}

func (o *WaterRecordOperator) handleCreateRecord(ctx context.Context, r *http.Request, user *utils.User) (context.Context, *framework.OperatorResult) {
	var req WaterRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Invalid request body", "BadRequest", http.StatusBadRequest),
		}
	}

	// 验证输入
	if req.Amount <= 0 {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Amount must be positive", "BadRequest", http.StatusBadRequest),
		}
	}

	// 保存记录到数据库（这里简化为模拟）
	record := WaterRecordResponse{
		ID:        "record_" + time.Now().Format("20060102150405"),
		Amount:    req.Amount,
		Time:      req.Time,
		DrinkType: req.DrinkType,
	}

	return ctx, &framework.OperatorResult{
		Data: record,
	}
}

func (o *WaterRecordOperator) handleGetRecords(ctx context.Context, r *http.Request, user *utils.User) (context.Context, *framework.OperatorResult) {
	// 解析查询参数
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// 模拟从数据库获取记录（这里简化为返回模拟数据）
	records := []WaterRecordResponse{
		{
			ID:        "record_1",
			Amount:    300,
			Time:      time.Now().Add(-2 * time.Hour),
			DrinkType: "water",
		},
		{
			ID:        "record_2",
			Amount:    200,
			Time:      time.Now().Add(-1 * time.Hour),
			DrinkType: "tea",
		},
	}

	return ctx, &framework.OperatorResult{
		Data: records,
	}
}
