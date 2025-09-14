package operators

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zhanghuachuan/water-reminder/database"
	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/types"
)

// 使用 types.ReminderConfig 替代本地定义

type ReminderConfigOperator struct{}

func (o *ReminderConfigOperator) Name() string {
	return "reminder-config"
}

func init() {
	framework.RegisterOperator("reminder-config", &ReminderConfigOperator{})
}

func (o *ReminderConfigOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	user := ctx.Value("user").(*types.User)

	switch r.Method {
	case http.MethodGet:
		return o.handleGetConfig(ctx, w, r, user)
	case http.MethodPost, http.MethodPut:
		return o.handleUpdateConfig(ctx, w, r, user)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return ctx, nil
	}
}

func (o *ReminderConfigOperator) handleGetConfig(ctx context.Context, w http.ResponseWriter, r *http.Request, user *types.User) (context.Context, error) {
	// 从数据库获取用户配置
	var config types.ReminderConfig
	if err := database.GetDB().Where("user_id = ?", user.ID).First(&config).Error; err != nil {
		http.Error(w, "Config not found", http.StatusNotFound)
		return ctx, err
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
	return ctx, nil
}

func (o *ReminderConfigOperator) handleUpdateConfig(ctx context.Context, w http.ResponseWriter, r *http.Request, user *types.User) (context.Context, error) {
	var config types.ReminderConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return ctx, err
	}

	// 设置用户ID
	config.UserID = user.ID

	// 验证配置
	if config.Interval < 15 {
		http.Error(w, "Interval must be at least 15 minutes", http.StatusBadRequest)
		return ctx, nil
	}
	if config.DailyTarget <= 0 {
		http.Error(w, "Daily target must be positive", http.StatusBadRequest)
		return ctx, nil
	}

	// 保存配置到数据库
	if err := database.GetDB().Where("user_id = ?", user.ID).Save(&config).Error; err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return ctx, err
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
	return ctx, nil
}
