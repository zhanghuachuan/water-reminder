package framework

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"
)

// 确保导入包被使用（消除编译器警告）
var _ = json.Compact
var _ = os.ReadFile

type ExecutionResult struct {
	OperatorName string
	Status       ExecutionStatus
	Err          error
}

type OperatorConfig struct {
	Type   string
	Config map[string]interface{}
}

type ExecutionStatus int

const (
	StatusSuccess ExecutionStatus = iota
	StatusFailed
)

type ExecuteOptions struct {
	Parallel bool
	Timeout  time.Duration
	Request  *http.Request
	Response http.ResponseWriter
}

type OperatorFactory interface {
	Create(config map[string]interface{}) (Operator, error)
	Name() string
}

type ConfigValidator struct {
	factories map[string]OperatorFactory
}

type Scheduler struct {
	operators    map[string]Operator
	dependencies map[string][]string
}

func NewConfigValidator(factories map[string]OperatorFactory) *ConfigValidator {
	return &ConfigValidator{
		factories: factories,
	}
}

func (v *ConfigValidator) Validate(configPath string) (*Scheduler, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config struct {
		Operators map[string]struct {
			Type   string                 `json:"type"`
			Config map[string]interface{} `json:"config"`
		} `json:"operators"`
		Dependencies map[string]string `json:"dependencies"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// 验证算子工厂
	for name, opConfig := range config.Operators {
		if _, exists := v.factories[opConfig.Type]; !exists {
			return nil, fmt.Errorf("operator factory not found for %s", name)
		}
	}

	// 构建依赖图并检测环
	deps := make(map[string][]string)
	for from, to := range config.Dependencies {
		if _, exists := config.Operators[from]; !exists {
			return nil, fmt.Errorf("source operator %s not found", from)
		}
		if _, exists := config.Operators[to]; !exists {
			return nil, fmt.Errorf("target operator %s not found", to)
		}
		deps[from] = append(deps[from], to)
	}

	if err := detectCycles(deps, config.Operators); err != nil {
		return nil, err
	}

	return &Scheduler{
		dependencies: deps,
	}, nil
}

func detectCycles(deps map[string][]string, operators interface{}) error {
	inDegree := make(map[string]int)
	queue := make([]string, 0)

	// 初始化所有算子的入度
	// 通用类型处理
	operatorsMap := reflect.ValueOf(operators)
	if operatorsMap.Kind() != reflect.Map {
		return errors.New("invalid operators type")
	}

	var opNames []string
	for _, key := range operatorsMap.MapKeys() {
		opNames = append(opNames, key.String())
	}

	for _, opName := range opNames {
		inDegree[opName] = 0
	}

	// 计算初始入度
	for _, toList := range deps {
		for _, to := range toList {
			inDegree[to]++
		}
	}

	// 找出入度为0的算子
	for opName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, opName)
		}
	}

	processed := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		processed++

		for _, neighbor := range deps[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if processed != len(opNames) {
		return errors.New("cycle detected in operator dependencies")
	}
	return nil
}

func NewScheduler(dependencies map[string][]string) *Scheduler {
	return &Scheduler{
		dependencies: dependencies,
	}
}

func (s *Scheduler) detectCycles(deps map[string][]string, operators map[string]struct{}) error {
	inDegree := make(map[string]int)
	queue := make([]string, 0)

	// 初始化所有算子的入度
	for opName := range operators {
		inDegree[opName] = 0
	}

	// 计算初始入度
	for _, toList := range deps {
		for _, to := range toList {
			inDegree[to]++
		}
	}

	// 找出入度为0的算子
	for opName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, opName)
		}
	}

	processed := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		processed++

		for _, neighbor := range deps[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if processed != len(operators) {
		return errors.New("cycle detected in operator dependencies")
	}
	return nil
}

// AddDependency 已弃用，依赖关系应在验证阶段通过ConfigValidator设置
func (s *Scheduler) AddDependency(from, to string) error {
	return errors.New("dependencies should be set during config validation phase")
}

func (s *Scheduler) Execute(ctx context.Context, factories map[string]OperatorFactory, opts ...ExecuteOptions) (map[string]ExecutionResult, error) {
	// 创建算子实例
	s.operators = make(map[string]Operator)
	for name := range s.dependencies {
		factory, exists := factories[name]
		if !exists {
			return nil, fmt.Errorf("operator factory not found for %s", name)
		}
		op, err := factory.Create(nil)
		if err != nil {
			return nil, err
		}
		s.operators[name] = op
	}

	// 2. 获取已排序的算子列表 (依赖关系已在LoadConfig验证)
	sortedOps, err := s.topologicalSort()
	if err != nil {
		return nil, fmt.Errorf("unexpected error in execution: %w", err)
	}

	options := ExecuteOptions{}
	if len(opts) > 0 {
		options = opts[0]
	}

	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	results := make(map[string]ExecutionResult)
	currentCtx := ctx

	for _, opName := range sortedOps {
		op := s.operators[opName]
		_, err := op.Execute(currentCtx, options.Response, options.Request)

		result := ExecutionResult{
			OperatorName: opName,
			Status:       StatusSuccess,
			Err:          err,
		}
		if err != nil {
			result.Status = StatusFailed
		}
		results[opName] = result

		if err != nil {
			break
		}
	}

	return results, nil
}

func (s *Scheduler) topologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	queue := make([]string, 0)
	result := make([]string, 0)

	for opName := range s.operators {
		inDegree[opName] = 0
	}

	for _, toList := range s.dependencies {
		for _, to := range toList {
			inDegree[to]++
		}
	}

	for opName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, opName)
		}
	}

	processed := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		processed++

		for _, neighbor := range s.dependencies[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] < 0 {
				return nil, errors.New("cycle detected in operator dependencies")
			}
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if processed != len(s.operators) {
		return nil, errors.New("cycle detected in operator dependencies")
	}

	return result, nil
}
