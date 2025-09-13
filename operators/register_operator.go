package operators

import (
	"context"
	"net/http"

	"github.com/zhanghuachuan/water-reminder/framework/types"
)

func init() {
	types.GetFactory().Register("user_register", &RegisterOperator{})
}

type RegisterOperator struct{}

func (o *RegisterOperator) Name() string {
	return "user_register"
}

func (o *RegisterOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	// 1. 解析请求数据
	// 2. 验证数据
	// 3. 创建用户
	// 4. 返回响应

	// 示例：设置用户ID到上下文
	return context.WithValue(ctx, "user_id", "new_user_id"), nil
}
