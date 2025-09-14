package framework

import (
	"context"
	"net/http"
)

// OperatorResult 算子执行结果
type OperatorResult struct {
	Data  interface{} // 成功时的业务数据
	Error error       // 错误信息
}

// Operator 算子接口
type Operator interface {
	Name() string
	Execute(ctx context.Context, r *http.Request) (context.Context, *OperatorResult)
}

// OperatorFactory 算子工厂接口
type OperatorFactory interface {
	Name() string
	Create(config map[string]interface{}) (Operator, error)
}
