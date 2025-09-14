package operators

import (
	"github.com/zhanghuachuan/water-reminder/framework"
)

var (
	initialized = false
)

// init 自动注册算子
func init() {
	InitializeOperators()
}

// InitializeOperators 显式初始化所有算子
func InitializeOperators() {
	if initialized {
		return
	}

	// 注册核心算子
	framework.RegisterOperator("validate", &ValidatorOperator{})
	framework.RegisterOperator("auth", &AuthOperator{})
	framework.RegisterOperator("register", &RegisterOperator{})
	framework.RegisterOperator("login", &LoginOperator{})
	framework.RegisterOperator("drinking-record", &DrinkingRecordOperator{})
	framework.RegisterOperator("statistics", &StatisticsOperator{})

	initialized = true
}
