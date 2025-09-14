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

func (o *ReminderConfigOperator) Execute(ctx context.Context, r *http.Request) (context.Context, *framework.OperatorResult) {
	user := ctx.Value("user").(*types.User)

	switch r.Method {
	case http.MethodGet:
		return o.handleGetConfig(ctx, r, user)
	case http.MethodPost, http.MethodPut:
		return o.handleUpdateConfig(ctx, r, user)
	default:
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Method not allowed", "MethodNotAllowed", http.StatusMethodNotAllowed),
		}
	}
}

func (o *ReminderConfigOperator) handleGetConfig(ctx context.Context, r *http.Request, user *types.User) (context.Context, *framework.OperatorResult) {
	// 从数据库获取用户配置
	var config types.ReminderConfig
	if err := database.GetDB().Where("user_id = ?", user.ID).First(&config).Error; err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Config not found", "NotFound", http.StatusNotFound),
		}
	}

	return ctx, &framework.OperatorResult{
		Data: config,
	}
}

func (o *ReminderConfigOperator) handleUpdateConfig(ctx context.Context, r *http.Request, user *types.User) (context.Context, *framework.OperatorResult) {
	var config types.ReminderConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Invalid request", "BadRequest", http.StatusBadRequest),
		}
	}

	// 设置用户ID
	config.UserID = user.ID

	// 验证配置
	if config.Interval < 15 {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Interval must be at least 15 minutes", "BadRequest", http.StatusBadRequest),
		}
	}
	if config.DailyTarget <= 0 {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Daily target must be positive", "BadRequest", http.StatusBadRequest),
		}
	}

	// 保存配置到数据库
	if err := database.GetDB().Where("user_id = ?", user.ID).Save(&config).Error; err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Failed to save config", "InternalServerError", http.StatusInternalServerError),
		}
	}

	return ctx, &framework.OperatorResult{
		Data: config,
	}
}
