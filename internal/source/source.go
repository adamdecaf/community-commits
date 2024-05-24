package source

func New(config Config) (Client, error) {
	switch {
	case config.Github != nil:
		return newGithubClient(*config.Github)
	}
	return nil, nil
}
