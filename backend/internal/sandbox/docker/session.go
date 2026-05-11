// the generic "create a container, run it, capture output, kill

package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// truncateString cuts a string off after max bytes. We use it on stdout/stderr
// so a runaway `while True: print(...)` doesn't fill up our DB.
func truncateString(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "\n... [truncated]"
}

// containerRunAndWait is the workhorse. It:
//
//  1. Creates the container with the given config.
//  2. Attaches to it so we can stream stdin in and pull stdout/stderr out.
//  3. Starts the container.
//  4. If we have stdin, spawns a goroutine that writes it then closes the pipe.
//  5. Spawns another goroutine that reads stdout+stderr until EOF using stdcopy
//  6. Races a timer against the read finishing. If the timer wins, we SIGKILL
//     and still wait for the reader so we don't leak output.
//  7. Inspects the container to grab the exit code and the OOMKilled flag.
//  8. Returns everything. The defer removes the container so it doesn't pile up.
//
// We use context.Background() for cleanup paths on purpose: even if the caller's  context got cancelled, we still want to kill/remove the container.

func containerRunAndWait(
	ctx context.Context,
	cli *client.Client,
	cfg *container.Config,
	hostCfg *container.HostConfig,
	attachOpts container.AttachOptions,
	stdinPayload string,
	timeoutMs int,
) (stdout, stderr string, exitCode int, oom bool, timedOut bool, elapsed time.Duration, err error) {
	created, err := cli.ContainerCreate(ctx, cfg, hostCfg, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return "", "", 0, false, false, 0, fmt.Errorf("container create: %w", err)
	}
	cid := created.ID
	defer func() {
		_ = cli.ContainerRemove(context.Background(), cid, container.RemoveOptions{Force: true})
	}()

	attach, err := cli.ContainerAttach(ctx, cid, attachOpts)
	if err != nil {
		return "", "", 0, false, false, 0, fmt.Errorf("container attach: %w", err)
	}
	defer attach.Close()

	if err := cli.ContainerStart(ctx, cid, container.StartOptions{}); err != nil {
		return "", "", 0, false, false, 0, fmt.Errorf("container start: %w", err)
	}

	start := time.Now()

	if attachOpts.Stdin {
		go func() {
			if stdinPayload != "" {
				_, _ = io.Copy(attach.Conn, bytes.NewBufferString(stdinPayload))
			}
			_ = attach.CloseWrite()
		}()
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		_, _ = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, attach.Reader)
		close(readDone)
	}()

	timer := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-readDone:
		// program finished on its own, good case.
	case <-timer.C:
		// hit the wall clock timeout — TLE in judge terms.
		timedOut = true
		_ = cli.ContainerKill(context.Background(), cid, "SIGKILL")
		<-readDone
	case <-ctx.Done():
		// caller cancelled (server shutting down, request cancelled, etc).
		_ = cli.ContainerKill(context.Background(), cid, "SIGKILL")
		<-readDone
		return "", "", 0, false, false, 0, ctx.Err()
	}

	elapsed = time.Since(start)

	// Inspect uses its own short context so a slow docker daemon during
	// shutdown can't hang us forever.
	inspectCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	inspect, _ := cli.ContainerInspect(inspectCtx, cid)
	if inspect.State != nil {
		exitCode = inspect.State.ExitCode
		oom = inspect.State.OOMKilled
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode, oom, timedOut, elapsed, nil
}
