package framework

import "sync"

// BaseOperatorFactory 基础算子工厂
type BaseOperatorFactory struct {
	name    string
	creator func(config map[string]interface{}) (Operator, error)
}

var (
	operatorFactories = make(map[string]OperatorFactory)
	factoryMutex      sync.RWMutex
)

// RegisterOperatorFactory 注册算子工厂
func RegisterOperatorFactory(name string, factory OperatorFactory) {
	factoryMutex.Lock()
	defer factoryMutex.Unlock()
	operatorFactories[name] = factory
}

// GetOperatorFactory 获取所有算子工厂
func GetOperatorFactory() map[string]OperatorFactory {
	factoryMutex.RLock()
	defer factoryMutex.RUnlock()

	// 返回副本避免并发修改
	factories := make(map[string]OperatorFactory)
	for name, factory := range operatorFactories {
		factories[name] = factory
	}
	return factories
}

func NewOperatorFactory(
	name string,
	creator func(config map[string]interface{}) (Operator, error),
) *BaseOperatorFactory {
	return &BaseOperatorFactory{
		name:    name,
		creator: creator,
	}
}

func (f *BaseOperatorFactory) Name() string {
	return f.name
}

func (f *BaseOperatorFactory) Create(config map[string]interface{}) (Operator, error) {
	return f.creator(config)
}
