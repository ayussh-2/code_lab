// builds the docker "HostConfig" the security + resource part of docker run.the sandboxing happens here
package docker

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
)

const (
	// containerUser is the uid:gid we run user code as inside the container.
	// Matches the `runner` user we created in the Dockerfiles (uid 1001).
	// NEVER run user code as root.
	containerUser = "1001:1001"

	// workdir is where we mount the per-submission folder inside the container.
	// All run/compile commands are relative to this.
	workdir = "/work"
)

// bindWorkMount builds the string Docker expects for the -v flag.
// readOnly=true is used for Run (user code shoudn't be able to overwrite the
// compiled binary or source). readOnly=false is used for Compile (compiler
// needs to write the artifact).
func bindWorkMount(hostArtifactPath string, readOnly bool) string {
	suffix := ":rw"
	if readOnly {
		suffix = ":ro"
	}
	return fmt.Sprintf("%s:%s%s", hostArtifactPath, workdir, suffix)
}

// newHostConfig builds the resource + security envelope for one container.
// Every flag here matters. Quick tour:
//   - NetworkMode "none": no network at all. Stops the program from phoning home.
//   - ReadonlyRootfs:     can't write anywhere except the tmpfs and the bind mount.
//   - CapDrop ALL:        drop every linux capability. No raw sockets, no mount, no ptrace.
//   - no-new-privileges:  even if a setuid binary exists, it can't elevate.
//   - Tmpfs /tmp:         a small writable scratch area. Sized so /tmp can't fill the host.
//   - Memory + MemorySwap equal: hard memory cap, no swap fallback.
//   - PidsLimit:          caps process count, blocks fork bombs.
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
		NetworkMode:    "none",
		ReadonlyRootfs: true,
		AutoRemove:     autoRemove,
		CapDrop:        []string{"ALL"},
		SecurityOpt:    []string{"no-new-privileges:true"},
		Tmpfs: map[string]string{
			"/tmp": fmt.Sprintf("size=%s,mode=1777", tmpfsSize),
		},
		Binds: []string{bind},
		Resources: container.Resources{
			Memory:     memBytes,
			MemorySwap: memBytes,
			NanoCPUs:   nanoCPUs,
			PidsLimit:  &p,
		},
	}
}
