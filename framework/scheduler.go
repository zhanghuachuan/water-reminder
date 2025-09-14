package framework

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

type ExecutionResult struct {
	OperatorName string
	Success      bool
	Error        error
	Duration     int64
	Data         interface{} // 存储算子执行结果的数据
}

type ExecuteOptions struct {
	Parallel bool
	Request  *http.Request
	Response http.ResponseWriter
}

type Scheduler struct {
	operators      map[string]Operator
	dependencies   map[string]map[string][]string // server_name -> dependencies
	executionOrder map[string][][]string          // server_name -> execution order (grouped by level)
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		dependencies:   make(map[string]map[string][]string),
		executionOrder: make(map[string][][]string),
	}
}

func (s *Scheduler) LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var configs []struct {
		ServerName   string            `json:"server_name"`
		Dependencies map[string]string `json:"dependencies"`
	}

	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	for _, config := range configs {
		// 将字符串依赖关系转换为字符串数组
		dependencies := make(map[string][]string)
		for op, dep := range config.Dependencies {
			dependencies[op] = []string{dep}
		}

		s.dependencies[config.ServerName] = dependencies
		if err := s.precomputeExecutionOrder(config.ServerName); err != nil {
			return fmt.Errorf("failed to precompute execution order for %s: %w", config.ServerName, err)
		}
	}

	return nil
}

func (s *Scheduler) precomputeExecutionOrder(serverName string) error {
	if s.operators == nil {
		s.operators = make(map[string]Operator)
	}

	dependencies := s.dependencies[serverName]
	for opName := range dependencies {
		// 使用全局算子注册表
		if op, err := GetOperator(opName); err == nil {
			s.operators[opName] = op
		} else {
			return fmt.Errorf("operator %s not found in registry", opName)
		}
	}

	executionOrder, err := s.topologicalSort(dependencies)
	if err != nil {
		return err
	}

	s.executionOrder[serverName] = executionOrder
	return nil
}

func (s *Scheduler) topologicalSort(dependencies map[string][]string) ([][]string, error) {
	inDegree := make(map[string]int)
	for op := range dependencies {
		inDegree[op] = 0
	}

	for _, deps := range dependencies {
		for _, dep := range deps {
			if dep != "" { // 忽略空字符串依赖
				inDegree[dep]++
			}
		}
	}

	queue := []string{}
	for op, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, op)
		}
	}

	result := [][]string{}
	for len(queue) > 0 {
		// 当前层级的所有算子可以并行执行
		currentLevel := make([]string, len(queue))
		copy(currentLevel, queue)
		result = append(result, currentLevel)

		nextQueue := []string{}
		for _, current := range queue {
			for _, dep := range dependencies[current] {
				if dep != "" { // 忽略空字符串依赖
					inDegree[dep]--
					if inDegree[dep] == 0 {
						nextQueue = append(nextQueue, dep)
					}
				}
			}
		}
		queue = nextQueue
	}

	// 计算实际处理的算子数量（排除空依赖）
	processedCount := 0
	for _, level := range result {
		processedCount += len(level)
	}

	if len(result) == 0 || processedCount != len(dependencies) {
		return nil, errors.New("cycle detected in dependencies")
	}

	return result, nil
}

func (s *Scheduler) Execute(ctx context.Context, serverName string, opts ExecuteOptions) ([]ExecutionResult, error) {
	executionOrder, exists := s.executionOrder[serverName]
	if !exists {
		return nil, fmt.Errorf("execution order not found for server: %s", serverName)
	}

	if opts.Parallel {
		return s.executeSmartParallel(ctx, executionOrder, opts)
	}
	return s.executeSequential(ctx, executionOrder, opts)
}

func (s *Scheduler) executeSequential(ctx context.Context, executionOrder [][]string, opts ExecuteOptions) ([]ExecutionResult, error) {
	results := []ExecutionResult{}
	for _, level := range executionOrder {
		for _, opName := range level {
			op, exists := s.operators[opName]
			if !exists {
				results = append(results, ExecutionResult{
					OperatorName: opName,
					Success:      false,
					Error:        fmt.Errorf("operator not found"),
				})
				continue
			}

			newCtx, result := op.Execute(ctx, opts.Request)
			results = append(results, ExecutionResult{
				OperatorName: opName,
				Success:      result.Error == nil,
				Error:        result.Error,
				Data:         result.Data,
			})

			if result.Error != nil {
				log.Printf("Operator %s failed: %v", opName, result.Error)
				return results, fmt.Errorf("operator %s failed: %w", opName, result.Error)
			}
			ctx = newCtx
		}
	}

	return results, nil
}

func (s *Scheduler) executeSmartParallel(ctx context.Context, executionOrder [][]string, opts ExecuteOptions) ([]ExecutionResult, error) {
	results := []ExecutionResult{}

	for _, level := range executionOrder {
		levelResults := make([]ExecutionResult, len(level))
		levelCtx := ctx // 每个层级共享同一个初始上下文

		// 并行执行当前层级的所有算子
		type opResult struct {
			idx      int
			result   ExecutionResult
			newCtx   context.Context
			hasError bool
		}

		resultCh := make(chan opResult, len(level))

		for i, opName := range level {
			go func(idx int, name string) {
				op, exists := s.operators[name]
				if !exists {
					resultCh <- opResult{
						idx: idx,
						result: ExecutionResult{
							OperatorName: name,
							Success:      false,
							Error:        fmt.Errorf("operator not found"),
						},
						hasError: true,
					}
					return
				}

				newCtx, operatorResult := op.Execute(levelCtx, opts.Request)
				execResult := ExecutionResult{
					OperatorName: name,
					Success:      operatorResult.Error == nil,
					Error:        operatorResult.Error,
					Data:         operatorResult.Data,
				}

				resultCh <- opResult{
					idx:      idx,
					result:   execResult,
					newCtx:   newCtx,
					hasError: operatorResult.Error != nil,
				}
			}(i, opName)
		}

		// 收集当前层级的所有结果
		var levelErrors []error
		successfulCtxs := []context.Context{}

		for i := 0; i < len(level); i++ {
			res := <-resultCh
			levelResults[res.idx] = res.result

			if res.hasError {
				levelErrors = append(levelErrors, res.result.Error)
				log.Printf("Operator %s failed: %v", res.result.OperatorName, res.result.Error)
			} else if res.newCtx != nil {
				successfulCtxs = append(successfulCtxs, res.newCtx)
			}
		}

		// 检查当前层级是否有错误
		if len(levelErrors) > 0 {
			return append(results, levelResults...), fmt.Errorf("parallel execution failed at level: %v", levelErrors)
		}

		// 收集当前层级的结果
		results = append(results, levelResults...)

		// 更新上下文：合并所有成功算子的上下文
		if len(successfulCtxs) > 0 {
			ctx = s.mergeContexts(successfulCtxs)
		}
	}

	return results, nil
}

// mergeContexts 合并多个上下文，优先保留后出现的值
func (s *Scheduler) mergeContexts(ctxs []context.Context) context.Context {
	if len(ctxs) == 0 {
		return context.Background()
	}

	merged := ctxs[0]
	for i := 1; i < len(ctxs); i++ {
		merged = s.mergeTwoContexts(merged, ctxs[i])
	}
	return merged
}

// mergeTwoContexts 合并两个上下文
func (s *Scheduler) mergeTwoContexts(ctx1, ctx2 context.Context) context.Context {
	merged := context.Background()

	// 从ctx1复制所有值
	if ctx1 != nil {
		if user := ctx1.Value("user"); user != nil {
			merged = context.WithValue(merged, "user", user)
		}
		// 可以添加其他需要合并的上下文键值
	}

	// 从ctx2复制所有值（会覆盖ctx1中的相同键）
	if ctx2 != nil {
		if user := ctx2.Value("user"); user != nil {
			merged = context.WithValue(merged, "user", user)
		}
		// 可以添加其他需要合并的上下文键值
	}

	return merged
}
