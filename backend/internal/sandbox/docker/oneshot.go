package docker

import (
	"context"
	"fmt"
	"math"

	"github.com/docker/docker/api/types/container"
)

const compileOutputMaxBytes = 8192

type oneShotInput struct {
	image     string
	cmd       []string
	hostDir   string
	bindRW    bool
	timeoutMs int
	memoryMB  int
	cpus      float64
	pidsLimit int
}

func (r *Runner) runOneShot(ctx context.Context, in oneShotInput) (output string, exitCode int, timedOut bool, err error) {
	mount := bindWorkMount(in.hostDir, !in.bindRW)
	fmt.Print(mount)
	nano := int64(math.Round(in.cpus * 1e9))
	hostCfg := newHostConfig(mount, in.memoryMB, in.pidsLimit, nano, "64m", false)

	cfg := &container.Config{
		Image:           in.image,
		Cmd:             in.cmd,
		WorkingDir:      workdir,
		User:            containerUser,
		NetworkDisabled: true,
		AttachStdout:    true,
		AttachStderr:    true,
		Tty:             false,
	}

	attachOpts := container.AttachOptions{
		Stream: true, Stdout: true, Stderr: true,
	}

	stdout, stderr, exitCode, _, timedOut, _, err := containerRunAndWait(
		ctx, r.cli, cfg, hostCfg, attachOpts, "", in.timeoutMs,
	)
	if err != nil {
		return "", 0, false, err
	}

	// Most compilers put errors on stderr but warnings on stdout, so we glue them together so the user sees the whole picture.
	combined := stdout
	if stderr != "" {
		if combined != "" {
			combined += "\n"
		}
		combined += stderr
	}
	return truncateString(combined, compileOutputMaxBytes), exitCode, timedOut, nil
}
