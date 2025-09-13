package framework

// BaseOperatorFactory 基础算子工厂
type BaseOperatorFactory struct {
	name    string
	creator func(config map[string]interface{}) (Operator, error)
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
