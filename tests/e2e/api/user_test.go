package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofrs/uuid/v5"

	"github.com/golang-cz/skeleton/data"
	"github.com/golang-cz/skeleton/proto"
)

func TestUser(t *testing.T) {
	id := uuid.Must(uuid.FromString("018c1b46-78c4-7484-b677-eb177d0b8f61"))
	user := &data.User{
		User: &proto.User{
			ID:        id,
			Email:     "jimmy.page@yardbirds.com",
			Firstname: "Jimmy",
			Lastname:  "Page",
		},
	}
	if err := E2E.DB.Save(user); err != nil {
		t.Fatalf("save user to DB: %v", err)
	}

	userOut, err := E2E.RPCClient.GetUser(context.Background(), id.String())
	if err != nil {
		t.Fatalf("load user from RPC: %v", err)
	}

	fmt.Println(userOut)
}
