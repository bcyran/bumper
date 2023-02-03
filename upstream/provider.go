package upstream

// VersionProvider tries to find the latest software version based on its source URL.
type VersionProvider interface {
	LatestVersion() (Version, error)
	Equal(other interface{}) bool
}

// NewVersionProvider tries to create a VersionProvider instance for a given URL.
// Returns nil if there's no suitable provider.
func NewVersionProvider(url string) VersionProvider {
	if gitHubProvider := newGitHubProvider(url); gitHubProvider != nil {
		return gitHubProvider
	}
	if gitLabProvider := newGitLabProvider(url); gitLabProvider != nil {
		return gitLabProvider
	}
	return nil
}
