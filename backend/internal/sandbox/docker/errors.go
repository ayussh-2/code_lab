// errors.go re-exports the shared sandbox sentinel errors so callers inside
// the docker package can keep using the short names without an extra import.
// The actual values live in the parent sandbox package so non-docker backends
// can return the same errors.
package docker

import "github.com/ayussh-2/internal/sandbox"

var (
	ErrUnknownLanguage = sandbox.ErrUnknownLanguage
	ErrCompileFailed   = sandbox.ErrCompileFailed
	ErrCompileTimeout  = sandbox.ErrCompileTimeout
)
