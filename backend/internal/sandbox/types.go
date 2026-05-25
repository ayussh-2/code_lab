package sandbox

import "context"

type Limits struct {
	RunTimeoutMs     int
	CompileTimeoutMs int
	MemoryMB         int
	CPUs             int
	PidsLimit        int
	StdoutMaxBytes   int
	StderrMaxBytes   int
}

type RunResult struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	DurationMs int
	MemoryKB   int
	TimedOut   bool
	// OutOfMemory is true if the kernel killed the container for using too much memory.
	OOM bool
}

type Runner interface {

	// prepare-workspace -> if(compiled-lang) build the binary else just return -> artifactId -> Run and Cleanup if error then error returned
	Compile(ctx context.Context, lang, source string) (artifactID string, output string, err error)

	// runs using the artifactId -> one per test case
	Run(ctx context.Context, lang, artifactID, stdin string, limits Limits) (RunResult, error)

	// deletes the workspace
	Cleanup(artifactID string) error
}
