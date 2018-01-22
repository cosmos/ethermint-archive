package version

// Major version component of the current release
const Major = 0

// Minor version component of the current release
const Minor = 5

// Fix version component of the current release
const Fix = 3

var (
	// Version is the full version string
	Version = "0.5.3"

	// GitCommit is set with --ldflags "-X main.gitCommit=$(git rev-parse --short HEAD)"
	GitCommit string
)

func init() {
	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}
