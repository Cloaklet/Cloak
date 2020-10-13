package version

// These variables are set during building
var (
	Version   string // the version string
	BuildTime string // build time in UTC
	GitCommit string // the git commit ID the app was built against
)
