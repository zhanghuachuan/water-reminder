package framework

import "errors"

var operators = make(map[string]Operator)

func RegisterOperator(name string, op Operator) {
	operators[name] = op
}

func GetOperator(name string) (Operator, error) {
	if op, exists := operators[name]; exists {
		return op, nil
	}
	return nil, errors.New("operator not found")
}
