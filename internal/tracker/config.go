package tracker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adamdecaf/community-commits/internal/source"

	"github.com/spf13/viper"
)

func Load(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("no path specified")
	}

	fullpath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("path %s expansion failed: %v", path, err)
	}

	var cfg Config

	fd, err := os.Open(fullpath)
	if err != nil {
		return nil, err
	}

	reader := viper.New()
	reader.SetConfigType("yaml")
	if err := reader.ReadConfig(fd); err != nil {
		return nil, err
	}
	if err := reader.UnmarshalExact(&cfg); err != nil {
		return nil, err
	}

	ReadSourcesFromEnv(&cfg.Sources)

	return &cfg, nil
}

type Config struct {
	Tracking TrackingConfig
	Sources  source.Config
}

type TrackingConfig struct {
	Repositories []Repository
}

type Repository struct {
	Source string
	Owner  string
	Name   string
}

func ReadSourcesFromEnv(existing *source.Config) {
	if v := os.Getenv("COMMUNITY_COMMITS_GITHUB_API_KEY"); v != "" {
		existing.Github = &source.GithubConfig{
			AuthToken: v,
		}
	}
}
