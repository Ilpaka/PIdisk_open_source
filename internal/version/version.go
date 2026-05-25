package version

var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func Full() string {
	return Version + " (" + Commit + " @ " + BuildTime + ")"
}
