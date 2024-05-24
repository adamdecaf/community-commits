package community_commits

import (
	"log/slog"
	"os"

	communitycommits "github.com/adamdecaf/community-commits"
)

func Setup() Environment {
	var env Environment

	env.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil)).
		With("app", "community_commits").
		With("version", communitycommits.Version())

	env.Logger.Info("starting adamdecaf/community-commits")

	return env
}
