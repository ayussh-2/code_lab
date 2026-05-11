// a "fire and forget" container run used for the compile step.
// Differences from a normal Run:
//   - no stdin
//   - workspace is mounted rw (compiler writes the binary into it)
//   - bigger /tmp because g++ likes scratch space
//   - stdout and stderr are merged into one string (we just want compiler logs)
package docker

import (
	"context"
	"math"

	"github.com/docker/docker/api/types/container"
)

// compileOutputMaxBytes caps how much compiler output we keep. 8K is way more
// than enough for "missing semicolon" style errors and small enough we can
// store it in the DB without a fuss.
const compileOutputMaxBytes = 8192

// oneShotInput collects everything runOneShot needs. Kept private since only
// Compile uses it.
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

// runOneShot starts a container, runs cmd, captures merged stdout+stderr, and
// returns the exit code + a timedOut flag. Used for compiling, but is generic
// enough that future "one off" tasks (eg a sanity check) can reuse it.
func (r *Runner) runOneShot(ctx context.Context, in oneShotInput) (output string, exitCode int, timedOut bool, err error) {
	mount := bindWorkMount(in.hostDir, !in.bindRW)
	// math.Round so cpus=0.5 doesnt get silently floored to 0 by int conversion.
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

	// Most compilers put errors on stderr but warnings on stdout, so we glue
	// them together so the user sees the whole picture.
	combined := stdout
	if stderr != "" {
		if combined != "" {
			combined += "\n"
		}
		combined += stderr
	}
	return truncateString(combined, compileOutputMaxBytes), exitCode, timedOut, nil
}
