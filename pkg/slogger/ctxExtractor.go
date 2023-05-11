package slogger

import (
	"context"

	"golang.org/x/exp/slog"
)

type ctxField string

// extractor of context values
func ctxExtractor(ctx context.Context, ctxField ctxField) (string, bool) {
	value, exists := ctx.Value(ctxField).(string)
	return value, exists
}

// Custom slog handler for extracting values from context
func (h *defaultHandler) Handle(ctx context.Context, r slog.Record) error {

	var ctxField ctxField = "vctraceid"
	slogField := "vctraceid"

	if myField, exists := ctxExtractor(ctx, ctxField); exists {
		r.AddAttrs(slog.String(string(slogField), myField))
	}

	return h.Handler.Handle(ctx, r)
}
