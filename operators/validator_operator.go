package operators

import (
	"context"
	"net/http"

	"github.com/zhanghuachuan/water-reminder/utils"
)

type ValidatorOperator struct{}

func (o *ValidatorOperator) Name() string {
	return "validate"
}

func (o *ValidatorOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	// 验证请求方法
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return ctx, nil
	}

	// 验证内容类型
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
		return ctx, nil
	}

	// 验证JWT令牌（如果存在）
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		tokenString := authHeader[len("Bearer "):]
		_, err := utils.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return ctx, err
		}
	}

	return ctx, nil
}
