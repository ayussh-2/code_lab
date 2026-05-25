package docker

import (
	"fmt"
	"os"
	"path/filepath"
)

// makes a fresh temp directory under baseWorkDir  writes the code into it and returns a path
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
