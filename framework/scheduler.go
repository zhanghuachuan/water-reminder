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

			newCtx, err := op.Execute(ctx, opts.Response, opts.Request)
			results = append(results, ExecutionResult{
				OperatorName: opName,
				Success:      err == nil,
				Error:        err,
			})

			if err != nil {
				log.Printf("Operator %s failed: %v", opName, err)
				return results, fmt.Errorf("operator %s failed: %w", opName, err)
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
		levelCtx := ctx // 每个层级共享同一个上下文

		// 并行执行当前层级的所有算子
		errCh := make(chan error, len(level))
		done := make(chan bool, len(level))
		ctxCh := make(chan context.Context, len(level))

		for i, opName := range level {
			go func(idx int, name string) {
				op, exists := s.operators[name]
				if !exists {
					levelResults[idx] = ExecutionResult{
						OperatorName: name,
						Success:      false,
						Error:        fmt.Errorf("operator not found"),
					}
					errCh <- fmt.Errorf("operator %s not found", name)
					done <- true
					return
				}

				newCtx, err := op.Execute(levelCtx, opts.Response, opts.Request)
				levelResults[idx] = ExecutionResult{
					OperatorName: name,
					Success:      err == nil,
					Error:        err,
				}

				if err != nil {
					log.Printf("Operator %s failed: %v", name, err)
					errCh <- fmt.Errorf("operator %s failed: %w", name, err)
				} else {
					ctxCh <- newCtx // 传递更新后的上下文
				}
				done <- true
			}(i, opName)
		}

		// 等待当前层级所有算子完成
		for i := 0; i < len(level); i++ {
			<-done
		}

		close(errCh)
		close(done)
		close(ctxCh)

		// 检查当前层级是否有错误
		var levelErrors []error
		for err := range errCh {
			levelErrors = append(levelErrors, err)
		}

		if len(levelErrors) > 0 {
			return append(results, levelResults...), fmt.Errorf("parallel execution failed at level: %v", levelErrors)
		}

		// 收集当前层级的结果
		results = append(results, levelResults...)

		// 更新上下文：使用最后一个成功的算子的上下文
		for newCtx := range ctxCh {
			ctx = newCtx
		}
	}

	return results, nil
}
