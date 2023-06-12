package lg

import "errors"

// errorCause recursively unwraps given error and returns the topmost
// non-nil error cause, same as github.com/pkg/errorCause(err).
func ErrorCause(err error) error {
	var cause error
	for e := err; e != nil; e = errors.Unwrap(e) {
		cause = e
	}
	return cause
}
