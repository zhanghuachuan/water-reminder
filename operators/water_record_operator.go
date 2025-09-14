package operators

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zhanghuachuan/water-reminder/framework"
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

func (o *WaterRecordOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	// 获取当前用户
	user, ok := ctx.Value("user").(*utils.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return ctx, nil
	}

	switch r.Method {
	case http.MethodPost:
		return o.handleCreateRecord(ctx, w, r, user)
	case http.MethodGet:
		return o.handleGetRecords(ctx, w, r, user)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return ctx, nil
	}
}

func (o *WaterRecordOperator) handleCreateRecord(ctx context.Context, w http.ResponseWriter, r *http.Request, user *utils.User) (context.Context, error) {
	var req WaterRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return ctx, err
	}

	// 验证输入
	if req.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return ctx, nil
	}

	// 保存记录到数据库（这里简化为模拟）
	record := WaterRecordResponse{
		ID:        "record_" + time.Now().Format("20060102150405"),
		Amount:    req.Amount,
		Time:      req.Time,
		DrinkType: req.DrinkType,
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)

	return ctx, nil
}

func (o *WaterRecordOperator) handleGetRecords(ctx context.Context, w http.ResponseWriter, r *http.Request, user *utils.User) (context.Context, error) {
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

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)

	return ctx, nil
}
