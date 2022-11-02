package version

type VersionProvider interface {
	LatestVersion() (string, error)
}

func NewVersionProvider(url string) VersionProvider {
	return NewGitHubProvider(url)
}
