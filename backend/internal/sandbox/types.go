// Package sandbox holds the shared types used by anything that runs untrustred
// user code. Other packages (like sandbox/docker) implement the Runner here.
// The whole point is: the rest of the app only talks to this interface so we
// can swap docker out for somthing else later if we want.
package sandbox

import "context"

// Limits are the resource caps + i/o caps we give to a single program run.
// All values come from config so the judge can tune them per language/problem.
type Limits struct {
	// RunTimeoutMs is the wall clock time we let the user code run before
	// we SIGKILL it. Used for the "TLE" verdict.
	RunTimeoutMs int

	// CompileTimeoutMs is how long the compile step (g++, etc) is allowed.
	// Separate from run timeout becuase compiling is allowed to be slower.
	CompileTimeoutMs int

	// MemoryMB is the hard memory ceiling. Going over this kills the container
	// with OOMKilled = true (we map that to "MLE").
	MemoryMB int

	// CPUs is fractional cpu cores (eg 0.5). int for now, switch to float64
	// if you ever want less than 1 core.
	CPUs int

	// PidsLimit caps how many processes the user code can spawn. Stops fork
	// bombs from eating the host.
	PidsLimit int

	// StdoutMaxBytes / StderrMaxBytes are the size at which we cut off the
	// captured output. Keeps gigantic prints from blowing up the DB.
	StdoutMaxBytes int
	StderrMaxBytes int
}

// RunResult is everything we learned from running the user code once.
// One of these per test case.
type RunResult struct {
	// Stdout is the trimed stdout of the program, capped at limits.StdoutMaxBytes.
	Stdout string
	// Stderr same as stdout but for stderr.
	Stderr string
	// ExitCode is the process exit code. 0 = normal exit.
	ExitCode int
	// DurationMs is how long the run actually took (wall clock).
	DurationMs int
	// MemoryKB is the peak memory used. Not implemented yet, will stay 0.
	MemoryKB int
	// TimedOut is true if we had to kill the container because it ran too long.
	TimedOut bool
	// OOM is true if the kernel killed the container for using too much memory.
	OOM bool
}

// Runner is the contract any sandbox backend has to satisfy.
// Right now only sandbox/docker implements this.
type Runner interface {
	// Compile prepares a workspace for the given source code. For compiled
	// languages it actually builds the binary. Returns an artifactID which
	// is just a handle (host dir path in the docker impl) that you pass back
	// to Run and Cleanup. The second string is the compiler output (only
	// usefull when err is ErrCompileFailed).
	Compile(ctx context.Context, lang, source string) (artifactID string, output string, err error)

	// Run executes the prepared artifact with the given stdin and limits.
	// One call per test case.
	Run(ctx context.Context, lang, artifactID, stdin string, limits Limits) (RunResult, error)

	// Cleanup deletes the workspace. Always call this in a defer once you're
	// done with the submission, otherwise temp dirs will pile up.
	Cleanup(artifactID string) error
}
