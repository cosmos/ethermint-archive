package version

// Major version component of the current release
const Major = 0

// Minor version component of the current release
const Minor = 2

// Fix version component of the current release
const Fix = 0

var (
	// Version is the full version string
	Version = "0.2.0"

	// GitCommit is set with --ldflags "-X main.gitCommit=$(git rev-parse HEAD)"
	GitCommit string
)

func init() {
	if GitCommit != "" && len(GitCommit) >= 8 {
		Version += "-" + GitCommit[:8]
	}
}
