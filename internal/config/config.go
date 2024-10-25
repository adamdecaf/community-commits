package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adamdecaf/community-commits/internal/source"
	"github.com/adamdecaf/community-commits/internal/tracker"

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

	return &cfg, nil
}

type Config struct {
	Tracking tracker.Config
	Sources  source.Config
}
