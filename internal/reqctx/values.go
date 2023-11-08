package reqctx

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

var (
	userIdKey        ctxKey = "userId"
	applicationIdKey ctxKey = "applicationId"
)

type ctxKey string

func (k ctxKey) String() string {
	return fmt.Sprintf("context value %q", string(k))
}

// GetUserId returns userId from given context. Use userId.IsNil() to check for empty value.
func GetUserId(ctx context.Context) uuid.UUID {
	userId, _ := ctx.Value(userIdKey).(uuid.UUID)
	return userId
}

func SetUserId(ctx context.Context, userId uuid.UUID) context.Context {
	return context.WithValue(ctx, userIdKey, userId)
}

// GetApplicationId returns applicationId from given context. Use applicationId.IsNil() to check for empty value.
func GetApplicationId(ctx context.Context) uuid.UUID {
	applicationId, _ := ctx.Value(applicationIdKey).(uuid.UUID)
	return applicationId
}

func SetApplicationId(ctx context.Context, applicationId uuid.UUID) context.Context {
	return context.WithValue(ctx, applicationIdKey, applicationId)
}
