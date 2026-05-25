package docker

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
)

const (
	containerUser = "1001:1001"
	workdir       = "/work"
)

// builds the string Docker expects for the -v flag
func bindWorkMount(hostArtifactPath string, readOnly bool) string {
	suffix := ":rw"
	if readOnly {
		suffix = ":ro"
	}
	return fmt.Sprintf("%s:%s%s", hostArtifactPath, workdir, suffix)
}

func newHostConfig(
	bind string,
	memoryMB int,
	pidsLimit int,
	nanoCPUs int64,
	tmpfsSize string,
	autoRemove bool,
) *container.HostConfig {
	memBytes := int64(memoryMB) * 1024 * 1024
	p := int64(pidsLimit)
	return &container.HostConfig{
		NetworkMode:    "none", // no network calls
		ReadonlyRootfs: true,   // cant wirte anywhere except the the workdir
		AutoRemove:     autoRemove,
		CapDrop:        []string{"ALL"}, // disable all linux ability
		SecurityOpt:    []string{"no-new-privileges:true"},
		Tmpfs: map[string]string{
			"/tmp": fmt.Sprintf("size=%s,mode=1777", tmpfsSize), //a small writable area
		},
		Binds: []string{bind},
		Resources: container.Resources{
			Memory:     memBytes, // mem limit
			MemorySwap: memBytes,
			NanoCPUs:   nanoCPUs,
			PidsLimit:  &p, // process count
		},
	}
}
