package operators

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/types"
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

func (o *RegisterOperator) Execute(ctx context.Context, r *http.Request) (context.Context, *framework.OperatorResult) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("请求参数错误", "Invalid request body", http.StatusBadRequest),
		}
	}

	// 验证输入
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("参数验证失败", "All fields are required", http.StatusBadRequest),
		}
	}

	// 创建用户
	user, err := utils.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("用户创建失败", "Failed to create user: "+err.Error(), http.StatusInternalServerError),
		}
	}

	// 生成JWT令牌
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("令牌生成失败", "Failed to generate token", http.StatusInternalServerError),
		}
	}

	// 返回注册成功数据
	registerData := &types.LoginResponseData{
		Token: token,
		User:  user,
	}

	return context.WithValue(ctx, "user", user), &framework.OperatorResult{
		Data: registerData,
	}
}
