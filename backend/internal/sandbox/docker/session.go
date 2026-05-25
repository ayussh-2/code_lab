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

func truncateString(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "\n... [truncated]"
}

// starts and gives ip captures op and finally returns the results
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

	// pass in input vals
	if attachOpts.Stdin {
		go func() {
			if stdinPayload != "" {
				_, _ = io.Copy(attach.Conn, bytes.NewBufferString(stdinPayload))
			}
			_ = attach.CloseWrite()
		}()
	}

	// read stdoutput
	var stdoutBuf, stderrBuf bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		_, _ = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, attach.Reader)
		close(readDone)
	}()

	timer := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timer.Stop()

	// resovling go routine
	select {
	// program exited without issues
	case <-readDone:
		// tle
	case <-timer.C:
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
