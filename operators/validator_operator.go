package operators

import (
	"context"
	"net/http"

	"github.com/zhanghuachuan/water-reminder/framework"
	"github.com/zhanghuachuan/water-reminder/types"
	"github.com/zhanghuachuan/water-reminder/utils"
)

type ValidatorOperator struct{}

func (o *ValidatorOperator) Name() string {
	return "validate"
}

func (o *ValidatorOperator) Execute(ctx context.Context, r *http.Request) (context.Context, *framework.OperatorResult) {
	// 验证请求方法
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("方法不允许", "Method not allowed", http.StatusMethodNotAllowed),
		}
	}

	// 验证内容类型
	if r.Header.Get("Content-Type") != "application/json" {
		return ctx, &framework.OperatorResult{
			Error: types.NewApiError("不支持的媒体类型", "Unsupported media type", http.StatusUnsupportedMediaType),
		}
	}

	// 验证JWT令牌（如果存在）
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		tokenString := authHeader[len("Bearer "):]
		_, err := utils.ValidateJWT(tokenString)
		if err != nil {
			return ctx, &framework.OperatorResult{
				Error: types.NewApiError("未授权", "Unauthorized", http.StatusUnauthorized),
			}
		}
	}

	return ctx, &framework.OperatorResult{}
}
