package operators

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zhanghuachuan/water-reminder/database"
	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/types"
	"github.com/zhanghuachuan/water-reminder/utils"
)

type LoginOperator struct{}

func (o *LoginOperator) Name() string {
	return "login"
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (o *LoginOperator) Execute(ctx context.Context, r *http.Request) (context.Context, *framework.OperatorResult) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("请求参数错误", "Invalid request body", http.StatusBadRequest),
		}
	}

	// 验证用户凭据
	user, err := utils.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("认证失败", "Invalid credentials", http.StatusUnauthorized),
		}
	}

	// 生成JWT令牌
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("令牌生成失败", "Failed to generate token", http.StatusInternalServerError),
		}
	}

	// 存储token到Redis，设置24小时过期
	err = database.StoreTokenInRedis(user.ID, token, 24*time.Hour)
	if err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("令牌存储失败", "Failed to store token", http.StatusInternalServerError),
		}
	}

	// 返回登录成功数据
	loginData := &types.LoginResponseData{
		Token: token,
		User: &types.User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		},
	}

	return context.WithValue(ctx, "user", &types.User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		}), &framework.OperatorResult{
			Data: loginData,
		}
}
