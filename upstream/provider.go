package upstream

import (
	"go.uber.org/config"
)

// VersionProvider tries to find the latest software version based on its source URL.
type VersionProvider interface {
	LatestVersion() (Version, error)
	Equal(other interface{}) bool
}

// NewVersionProvider tries to create a VersionProvider instance for a given URL.
// Returns nil if there's no suitable provider.
func NewVersionProvider(url string, providersConfig config.Value) VersionProvider {
	if pypiProvider := newPypiProvider(url); pypiProvider != nil {
		return pypiProvider
	}
	if gitHubProvider := newGitHubProvider(url, providersConfig.Get("github")); gitHubProvider != nil {
		return gitHubProvider
	}
	if gitLabProvider := newGitLabProvider(url, providersConfig.Get("gitlab")); gitLabProvider != nil { // nolint:revive
		return gitLabProvider
	}
	return nil
}
