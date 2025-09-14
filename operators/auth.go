package operators

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/zhanghuachuan/water-reminder/database"
	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/types"
	"github.com/zhanghuachuan/water-reminder/utils"
)

type AuthResponse struct {
	UserID string `json:"userId"`
}

type AuthOperator struct{}

func (o *AuthOperator) Name() string {
	return "auth"
}

func (o *AuthOperator) Execute(ctx context.Context, r *http.Request) (context.Context, *framework.OperatorResult) {
	// 从Authorization头获取token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Authorization header required", "Unauthorized", http.StatusUnauthorized),
		}
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Invalid authorization format", "Unauthorized", http.StatusUnauthorized),
		}
	}

	// 验证JWT
	userID, err := utils.ValidateJWT(tokenString)
	if err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Invalid token", "Unauthorized", http.StatusUnauthorized),
		}
	}

	// 验证Redis中的token
	valid, err := database.IsTokenValid(userID, tokenString)
	if err != nil || !valid {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Invalid token", "Unauthorized", http.StatusUnauthorized),
		}
	}

	// 续期token（如果当天已经验证过，延长至一天）
	err = database.RefreshTokenInRedis(userID, 24*time.Hour)
	if err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("Failed to refresh token", "InternalServerError", http.StatusInternalServerError),
		}
	}

	// 将用户信息存入上下文
	ctx = context.WithValue(ctx, "user", &types.User{ID: userID})

	return ctx, &framework.OperatorResult{
		Data: AuthResponse{UserID: userID},
	}
}
