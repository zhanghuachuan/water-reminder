package framework

import (
	"testing"
)

func TestSchedulerErrorCases(t *testing.T) {
	factories := map[string]OperatorFactory{
		"op1": NewOperatorFactory("op1", func(config map[string]interface{}) (Operator, error) {
			return NewMockOperator("op1"), nil
		}),
		"op2": NewOperatorFactory("op2", func(config map[string]interface{}) (Operator, error) {
			return NewMockOperator("op2"), nil
		}),
	}

	t.Run("环检测", func(t *testing.T) {
		validator := NewConfigValidator(factories)
		_, err := validator.Validate(CreateTempConfig(map[string][]string{
			"op1": {"op2"},
			"op2": {"op1"},
		}))
		if err == nil {
			t.Error("预期检测到循环依赖，但未报错")
		}
	})

	t.Run("算子不存在", func(t *testing.T) {
		validator := NewConfigValidator(factories)
		_, err := validator.Validate(CreateTempConfig(map[string][]string{
			"nonexist": {"op1"},
		}))
		if err == nil {
			t.Error("预期检测到不存在的算子，但未报错")
		}
	})
}

// 使用test_utils.go中的CreateTempConfig
