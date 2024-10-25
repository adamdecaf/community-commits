package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	conf, err := Load(filepath.Join("testdata", "valid.yaml"))
	require.NoError(t, err)
	require.NotNil(t, conf)

	repos := conf.Tracking.Repositories
	require.Len(t, repos, 1)

	repo := repos[0]
	require.Equal(t, "github", repo.Source)
	require.Equal(t, "moov-io", repo.Owner)
	require.Equal(t, "ach", repo.Name)

	require.NotNil(t, conf.Sources.Github)
	require.Equal(t, "github-api-key", conf.Sources.Github.AuthToken)
}
