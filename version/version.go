package version

import (
	"fmt"

	"github.com/ethereum/go-ethereum/params"
)

const (
	Major = 0          // Major version component of the current release
	Minor = 1          // Minor version component of the current release
	Patch = 0          // Patch version component of the current release
	Meta  = "unstable" // Version metadata to append to the version string
)

var (
	// The full version string
	Version string

	// GitCommit is set with --ldflags "-X main.gitCommit=$(git rev-parse HEAD)"
	GitCommit string
)

func init() {
	Version = "ethermint/" + fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
	if Meta != "" {
		Version += "-" + Meta
	}

	if GitCommit != "" && len(GitCommit) >= 8 {
		Version += "-" + GitCommit[:8]
	}
	Version += " ++" + " go-ethereum/" + params.Version
}
