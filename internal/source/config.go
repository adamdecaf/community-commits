package source

type Config struct {
	Github *GithubConfig
}

type GithubConfig struct {
	AuthToken string
}
