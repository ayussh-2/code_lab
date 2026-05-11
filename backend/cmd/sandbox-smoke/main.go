// sandbox-smoke is a tiny binary that proves the sandbox runner works end to
// end without the rest of the judge pipeline in the way. Run it after building
// the sandbox images:
//
//	make sandbox-images
//	make sandbox-smoke
//
// What it does: takes the Python program `print(input())`, sends it the string
// "hello" on stdin, and checks the stdout matches "hello". If it does, prints
// "OK" and exits 0. If anything is off it exits 1 with a hopefully helpfull
// error message.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ayussh-2/internal/sandbox"
	"github.com/ayussh-2/internal/sandbox/docker"
)

func main() {
	// Outer context: 30s is plenty for one compile + one run.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cli, err := docker.NewDefaultClient()
	if err != nil {
		fail("docker client: %v", err)
	}
	defer cli.Close()

	// Use the OS temp dir so we don't litter the repo with workspace folders.
	workDir := filepath.Join(os.TempDir(), "codelab-sandbox-smoke")
	runner := docker.NewRunner(cli, workDir)

	source := "print(input())\n"
	artifactID, compileOut, err := runner.Compile(ctx, "python", source)
	if err != nil {
		fail("compile: %v\n%s", err, compileOut)
	}
	defer func() { _ = runner.Cleanup(artifactID) }()

	// These limits roughly mirror what the real judge will use. Tune in config
	// later, hardcoded here for the smoke test.
	limits := sandbox.Limits{
		RunTimeoutMs:     2000,
		CompileTimeoutMs: 10000,
		MemoryMB:         256,
		CPUs:             1,
		PidsLimit:        128,
		StdoutMaxBytes:   65536,
		StderrMaxBytes:   65536,
	}

	result, err := runner.Run(ctx, "python", artifactID, "hello\n", limits)
	if err != nil {
		fail("run: %v", err)
	}

	// Print everything we got back so debugging is easy when it breaks.
	fmt.Printf("stdout=%q\n", result.Stdout)
	fmt.Printf("stderr=%q\n", result.Stderr)
	fmt.Printf("exit_code=%d duration_ms=%d timed_out=%v oom=%v\n",
		result.ExitCode, result.DurationMs, result.TimedOut, result.OOM)

	// Each of these checks maps to a real verdict the judge will produce.
	if result.TimedOut {
		fail("smoke failed: timed out")
	}
	if result.OOM {
		fail("smoke failed: oom killed")
	}
	if result.ExitCode != 0 {
		fail("smoke failed: exit code %d", result.ExitCode)
	}
	if !sandbox.Equal("hello", result.Stdout) {
		fail("smoke failed: stdout mismatch (expected %q got %q)", "hello", result.Stdout)
	}

	fmt.Println("OK")
}

// fail prints to stderr and exits non-zero. Used everywhere instead of panic
// so the make target shows a clean error instead of a stack trace.
func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
