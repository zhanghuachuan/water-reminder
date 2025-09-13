package framework

import (
	"context"
	"testing"
)

func TestScheduler(t *testing.T) {
	factories := map[string]OperatorFactory{
		"op1": NewOperatorFactory("op1", func(config map[string]interface{}) (Operator, error) {
			return NewMockOperator("op1"), nil
		}),
		"op2": NewOperatorFactory("op2", func(config map[string]interface{}) (Operator, error) {
			return NewMockOperator("op2"), nil
		}),
	}

	validator := NewConfigValidator(factories)
	scheduler, err := validator.Validate(CreateTempConfig(map[string][]string{
		"op1": {"op2"},
		"op2": {}, // 确保无环
	}))
	if err != nil {
		t.Fatalf("配置验证失败: %v", err)
	}

	results, err := scheduler.Execute(context.Background(), factories)
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("预期执行2个算子，实际执行了%d个", len(results))
	}
}

// 使用test_utils.go中的CreateTempConfig
