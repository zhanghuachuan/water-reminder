package operators

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/utils"
)

func init() {
	framework.RegisterOperator("user_register", &RegisterOperator{})
}

type RegisterOperator struct{}

func (o *RegisterOperator) Name() string {
	return "register"
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (o *RegisterOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return ctx, err
	}

	// 验证输入
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return ctx, nil
	}

	// 创建用户
	user, err := utils.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return ctx, err
	}

	// 生成JWT令牌
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return ctx, err
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
	})

	return context.WithValue(ctx, "user", user), nil
}
