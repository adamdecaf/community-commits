package source

import (
	"fmt"
	"strings"
)

// ByName returns a source.Client for a given name (e.g. github, gitlab, etc)
func ByName(name string, config Config) (Client, error) {
	name = strings.ToLower(strings.TrimSpace(name))

	switch name {
	case "github":
		if config.Github != nil {
			return newGithubClient(*config.Github)
		}
	}

	return nil, fmt.Errorf("unknown %s source", name)
}
