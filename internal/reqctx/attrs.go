package reqctx

import (
	"context"
)

type ctxAttrKey struct{}

func (ctxAttrKey) String() string {
	return "reqctx attributes"
}

// NewAttrStorage creates a map storage for request-scoped attributes and
// returns new context holding its reference and the map itself.
//
// A typical use case: Create attr storage in request logger middleware, pass
// it down the chain via r.WithContext(ctx) and let any subsequent middlewares
// or handlers enrich the storage with new attributes.
//
// AddAttr effectively lets you bubble up values from any HTTP handler to the
// top-most middleware without having to write multiple log lines.
func NewAttrStorage(ctx context.Context) (context.Context, map[string]any) {
	m := map[string]any{}

	return context.WithValue(ctx, ctxAttrKey{}, m), m
}

// AddAttr adds a new attribute to the request context storage.
// It's not safe for concurrent access.
func AddAttr(ctx context.Context, key string, value any) {
	if m, ok := ctx.Value(ctxAttrKey{}).(map[string]any); ok {
		m[key] = value
	}
}
