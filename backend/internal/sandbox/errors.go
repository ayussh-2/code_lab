// errors.go: shared sentinel errors that any Runner implementation should
// return. Callers (the judge service) use errors.Is to decide what verdict to
// produce, so these belong at the abstraction level, not inside a specific
// backend like docker.
package sandbox

import "errors"

var (
	// ErrUnknownLanguage: the lang string didn't match any registered language.
	ErrUnknownLanguage = errors.New("unknown language")

	// ErrCompileFailed: compiler ran but exited non-zero. The judge service
	// maps this to a "CE" verdict and shows the compiler output to the user.
	ErrCompileFailed = errors.New("compile failed")

	// ErrCompileTimeout: compile step ran longer than the compile timeout.
	ErrCompileTimeout = errors.New("compile timed out")
)
