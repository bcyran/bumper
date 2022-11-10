package upstream

// VersionProvider tries to find the latest software version based on its source URL.
type VersionProvider interface {
	LatestVersion() (Version, error)
}

// NewVersionProvider tries to create a VersionProvider instance for a given URL.
// Returns nil if there's no suitable provider.
func NewVersionProvider(url string) VersionProvider {
	return newGitHubProvider(url)
}
