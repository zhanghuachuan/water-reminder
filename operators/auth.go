package operators

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

func (o *AuthOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	// 从Authorization头获取token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return ctx, nil
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
		return ctx, nil
	}

	// 验证JWT
	userID, err := utils.ValidateJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return ctx, err
	}

	// 验证Redis中的token
	valid, err := utils.IsTokenValid(userID, tokenString)
	if err != nil || !valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return ctx, err
	}

	// 续期token（如果当天已经验证过，延长至一天）
	err = utils.RefreshTokenInRedis(userID, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
		return ctx, err
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		UserID: userID,
	})

	// 将用户信息存入上下文
	return context.WithValue(ctx, "user", &types.User{ID: userID}), nil
}
