//go:generate go run github.com/golang-cz/gospeak/cmd/gospeak ./
package proto

import (
	"context"
)

//go:webrpc json -out=./docs/skeletonApi.webrpc.json
//go:webrpc golang@v0.13.5 -server -types=false -pkg=proto -errorStackTrace -fixEmptyArrays -out=./server.gen.go
//go:webrpc golang@v0.13.5 -client -pkg=skeleton -out=./client/skeleton/skeletonClient.gen.go
type Skeleton interface {
	Users
}

//go:webrpc openapi -title=SkeletonUsersAPI -serverUrl=https://dev.golang.cz/_api -out=./docs/skeletonUsersApi.gen.yaml
//go:webrpc typescript -client -out=./client/users/skeletonUsersClient.gen.ts
type Users interface {
	GetUser(ctx context.Context, id string) (user *User, err error)
}
