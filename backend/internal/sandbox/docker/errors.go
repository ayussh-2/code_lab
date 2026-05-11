// sentinel errors returned by the docker runner.
package docker

import "errors"

var (

	ErrUnknownLanguage = errors.New("unknown language")

	
	ErrCompileFailed = errors.New("compile failed")

	
	ErrCompileTimeout = errors.New("compile timed out")
)
