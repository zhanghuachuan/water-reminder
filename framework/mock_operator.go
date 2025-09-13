package framework

import (
	"context"
	"net/http"
)

type MockOperator struct {
	name string
}

func (m *MockOperator) Name() string {
	return m.name
}

func NewMockOperator(name string) *MockOperator {
	return &MockOperator{name: name}
}

func (m *MockOperator) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	return ctx, nil
}
