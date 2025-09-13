package operators

import (
	"context"
	"net/http"

	"github.com/zhanghuachuan/water-reminder/framework"
)

func init() {
	framework.GetFactory().Register("auth", &AuthOperator{})
}

type AuthOperator struct{}

func (o *AuthOperator) Name() string {
	return "auth"
}

func (o *AuthOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	// 认证逻辑实现
	return ctx, nil
}
