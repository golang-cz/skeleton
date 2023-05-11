package slogger

import (
	"context"

	"golang.org/x/exp/slog"
)

type CtxField string

var (
	firstCtxField CtxField = "firstCtxField"
)

// extractor of context values
func firstValueFromCtx(ctx context.Context) (string, bool) {
	value, exists := ctx.Value(firstCtxField).(string)
	return value, exists
}

// Custom slog handler for extracting values from context
func (h *DefaultHandler) Handle(ctx context.Context, r slog.Record) error {
	if myField, exists := firstValueFromCtx(ctx); exists {
		r.AddAttrs(slog.String(string(firstCtxField), myField))
	}

	return h.Handler.Handle(ctx, r)
}
