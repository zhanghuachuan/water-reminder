package operators

import (
	"context"
	"net/http"

	"github.com/zhanghuachuan/water-reminder/framework/types"
)

func init() {
	types.GetFactory().Register("request_validator", &ValidatorOperator{})
}

type ValidatorOperator struct{}

func (o *ValidatorOperator) Name() string {
	return "request_validator"
}

func (o *ValidatorOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	// 请求验证逻辑
	return ctx, nil
}
