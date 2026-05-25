package sandbox

import "errors"

var (
	ErrUnknownLanguage = errors.New("unknown language")

	// compile failed  / Compile Error(CE)
	ErrCompileFailed = errors.New("compile failed")

	ErrCompileTimeout = errors.New("compile timed out")
)
