package source

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGithub_GetForks(t *testing.T) {
	gh := testGithubClient(t)

	ctx := context.Background()
	forks, err := gh.GetForks(ctx, "moov-io", "ach")
	require.NoError(t, err)
	require.Greater(t, len(forks), 1)
}

func TestGithub_ListBranches(t *testing.T) {
	gh := testGithubClient(t)

	ctx := context.Background()
	branches, err := gh.ListBranches(ctx, Repository{
		Owner: "moov-io",
		Name:  "ach",
	})
	require.NoError(t, err)
	require.Greater(t, len(branches), 1)

	for i := range branches {
		t.Logf("%#v", branches[i])
	}
}

func testGithubClient(t *testing.T) *githubSource {
	t.Helper()

	if testing.Short() {
		t.Skip("-short flag provided")
	}

	authToken := os.Getenv("COMMUNITY_COMMITS_TEST_GITHUB_API_KEY")
	if authToken == "" {
		t.Skip("no Github ApiKey provided")
	}

	conf := GithubConfig{
		AuthToken: authToken,
	}
	cc, err := newGithubClient(conf)
	if err != nil {
		require.NoError(t, err)
	}
	return cc.(*githubSource)
}
