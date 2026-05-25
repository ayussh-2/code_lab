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

func NewRunner(cli *client.Client, baseWorkDir string) *Runner {
	return &Runner{cli: cli, baseWorkDir: baseWorkDir}
}

func NewDefaultClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

var _ sandbox.Runner = (*Runner)(nil)

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

	stdout, stderr, exitCode, oom, timedOut, elapsed, memoryKB, err := containerRunAndWait(
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
		MemoryKB:   memoryKB,
		TimedOut:   timedOut,
		OOM:        oom,
	}, nil
}

func (r *Runner) Cleanup(artifactID string) error {
	if artifactID == "" {
		return nil
	}
	return os.RemoveAll(artifactID)
}
