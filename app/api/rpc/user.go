package rpc

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"

	"github.com/golang-cz/skeleton/proto"
)

func (r *Rpc) GetUser(ctx context.Context, userId string) (*proto.User, error) {
	userUUUID, err := uuid.FromString(userId)
	if err != nil {
		return nil, fmt.Errorf("get uuid from string: %w", err)
	}

	user, err := r.DB.User.FindOne(userUUUID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user.User, nil
}
