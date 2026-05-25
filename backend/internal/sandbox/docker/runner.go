// Package docker is the docker-backed implementation of sandbox.Runner.
//
// Lifecycle of a submission:
//  1. Compile(lang, source)   -> creates host workspace, builds binary if needed.
//  2. Run(lang, artifact, ..) -> one call per test case, returns RunResult.
//  3. Cleanup(artifact)       -> deletes the host workspace. Always in defer.
//
// The host workspace is just a temp folder. Its path doubles as the "artifactID"
// we hand back to the caller, so the caller dosen't need to know how we store
// things.

package docker

import (
	"context"
	"errors"
	"os"

	"github.com/ayussh-2/internal/sandbox"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Runner struct {
	cli         *client.Client
	baseWorkDir string
}

// NewRunner wires up a Runner with a docker client and a base directory under
// which every submission gets its own subfolder (eg /tmp/codelab-sandbox/sub-*).
func NewRunner(cli *client.Client, baseWorkDir string) *Runner {
	return &Runner{cli: cli, baseWorkDir: baseWorkDir}
}

func NewDefaultClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// Compile time check that Runner satisfies the sandbox.Runner interface. If
// you change the interface and forget to update this struct the build breaks
// here, not somewhere deep in the judge service.
var _ sandbox.Runner = (*Runner)(nil)

// Compile prepares everything we need to run user code. For interpreted
// languages this just writes the source file into a workspace. For compiled
// langs it ALSO spins up a one-shot container that runs the compiler.
//
// Returns the artifact dir path, the compiler output (if any), and an error.
// Three failure modes worth distinguishing:
//   - err == ErrCompileFailed: compiler ran fine but exited non-zero. The
//     `output` string is the compiler stderr, which we want to show the user
//     as a "Compilation Error" verdict.
//   - err == ErrCompileTimeout: compile took longer than the compile timeout.
//   - any other err: infra problem, fail the submission with "Internal Error".
func (r *Runner) Compile(ctx context.Context, lang, source string) (string, string, error) {
	spec, ok := Languages[lang]
	if !ok {
		return "", "", ErrUnknownLanguage
	}

	artifactDir, err := createArtifactWorkspace(r.baseWorkDir, spec.FileName, source)
	if err != nil {
		return "", "", err
	}

	if !spec.NeedsCompile {
		// Interpreted lang, nothing else to do. The source is already in /work.
		return artifactDir, "", nil
	}

	output, exitCode, timedOut, err := r.runOneShot(ctx, oneShotInput{
		image:     spec.Image,
		cmd:       spec.CompileCmd,
		hostDir:   artifactDir,
		bindRW:    true,
		timeoutMs: 10000,
		memoryMB:  512,
		cpus:      1.0,
		pidsLimit: 256,
	})
	if err != nil {
		_ = os.RemoveAll(artifactDir)
		return "", output, err
	}
	if timedOut {
		_ = os.RemoveAll(artifactDir)
		return "", output, ErrCompileTimeout
	}
	if exitCode != 0 {

		return artifactDir, output, ErrCompileFailed
	}
	return artifactDir, output, nil
}

// Run executes the artifact once with the given stdin, capped by limits.
// One test case = one Run call. The workspace is mounted READ-ONLY here so a
// program can't tamper with its own binary or source for the next test.
func (r *Runner) Run(ctx context.Context, lang, artifactID, stdin string, limits sandbox.Limits) (sandbox.RunResult, error) {
	spec, ok := Languages[lang]
	if !ok {
		return sandbox.RunResult{}, ErrUnknownLanguage
	}
	if artifactID == "" {
		return sandbox.RunResult{}, errors.New("artifactID is required")
	}

	nano := int64(limits.CPUs) * 1_000_000_000
	mount := bindWorkMount(artifactID, true)
	hostCfg := newHostConfig(mount, limits.MemoryMB, limits.PidsLimit, nano, "16m", false)

	cfg := &container.Config{
		Image:           spec.Image,
		Cmd:             spec.RunCmd,
		WorkingDir:      workdir,
		User:            containerUser,
		NetworkDisabled: true,
		AttachStdin:     true,
		AttachStdout:    true,
		AttachStderr:    true,
		OpenStdin:       true,
		StdinOnce:       true,
		Tty:             false,
	}

	attachOpts := container.AttachOptions{
		Stream: true, Stdin: true, Stdout: true, Stderr: true,
	}

	stdout, stderr, exitCode, oom, timedOut, elapsed, err := containerRunAndWait(
		ctx, r.cli, cfg, hostCfg, attachOpts, stdin, limits.RunTimeoutMs,
	)
	if err != nil {
		return sandbox.RunResult{}, err
	}

	return sandbox.RunResult{
		Stdout:     truncateString(stdout, limits.StdoutMaxBytes),
		Stderr:     truncateString(stderr, limits.StderrMaxBytes),
		ExitCode:   exitCode,
		DurationMs: int(elapsed.Milliseconds()),
		TimedOut:   timedOut,
		OOM:        oom,
	}, nil
}

// Cleanup removes the per-submission workspace from the host.
func (r *Runner) Cleanup(artifactID string) error {
	if artifactID == "" {
		return nil
	}
	return os.RemoveAll(artifactID)
}
