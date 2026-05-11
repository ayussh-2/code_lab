// host filesystem helpers. The runner needs a folder per
// submission to put the source code (and compiled binary) into, and then bind
// mount that folder into the container. This file is the "make and fill that
// folder" part.
package docker

import (
	"fmt"
	"os"
	"path/filepath"
)

// createArtifactWorkspace makes a fresh temp directory under baseWorkDir,
// writes the user's source into it, and returns the path. The caller passes
// that path back to docker as the bind mount source.
//
// The chmod 0777 is on purpose: the container runs as uid 1001 (a different
// uid from whatever process we're in), and the easiest way to let it read
// (and for compile, write) into this folder is to open the perms wide. It's
// safe because each submission gets its own folder + container.
//
// On any failure after the dir is created we try to remove it so we don't
// leak temp dirs. Cleanup of succesfull workspaces happens later via
// Runner.Cleanup once judging is done.
func createArtifactWorkspace(baseWorkDir, fileName, source string) (artifactDir string, err error) {
	if err := os.MkdirAll(baseWorkDir, 0o755); err != nil {
		return "", fmt.Errorf("create base workdir: %w", err)
	}

	artifactDir, err = os.MkdirTemp(baseWorkDir, "sub-*")
	if err != nil {
		return "", fmt.Errorf("create artifact dir: %w", err)
	}

	if err := os.Chmod(artifactDir, 0o777); err != nil {
		_ = os.RemoveAll(artifactDir)
		return "", err
	}

	srcPath := filepath.Join(artifactDir, fileName)
	if err := os.WriteFile(srcPath, []byte(source), 0o666); err != nil {
		_ = os.RemoveAll(artifactDir)
		return "", fmt.Errorf("write source: %w", err)
	}

	return artifactDir, nil
}
