package operators

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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

func (o *LoginOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return ctx, err
	}

	// 验证用户凭据
	user, err := utils.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return ctx, err
	}

	// 生成JWT令牌
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return ctx, err
	}

	// 存储token到Redis，设置24小时过期
	err = utils.StoreTokenInRedis(user.ID, token, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to store token", http.StatusInternalServerError)
		return ctx, err
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": &types.User{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		},
	})

	return context.WithValue(ctx, "user", &types.User{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
	}), nil
}
