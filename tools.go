//go:build tools

package skeleton

// CLI tools used for the dev environment.
//
// Add a new tool to go.mod file via `go get <pkg>@latest`.
//
// You can then use each tool anywhere in the codebase using a go:generate
// directive, ie. // go:generate go run github.com/VojtechVitek/rerun/cmd/rerun <args>
//
// For more info, see https://gist.github.com/tschaub/66f5feb20ae1b5166e9fe928c5cba5e4.
import (
	_ "github.com/golang-cz/gospeak/cmd/gospeak"
	_ "github.com/mikefarah/yq/v4"
)
