package framework

import (
	"context"
	"net/http"
)

type Operator interface {
	Name() string
	Execute(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error)
}
