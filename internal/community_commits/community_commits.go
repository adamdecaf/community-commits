package community_commits

import (
	"fmt"
	"log/slog"
	"os"

	communitycommits "github.com/adamdecaf/community-commits"
	"github.com/adamdecaf/community-commits/internal/tracker"
)

func Setup(configFilepath string) (Environment, error) {
	var env Environment

	env.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil)).
		With("app", "community_commits").
		With("version", communitycommits.Version())

	env.Logger.Info("starting adamdecaf/community-commits")

	conf, err := tracker.Load(configFilepath)
	if err != nil {
		return env, fmt.Errorf("reading %s failed: %w", configFilepath, err)
	}
	env.Config = *conf

	if env.TrackingWorker == nil {
		w, err := tracker.NewWorker(env.Logger, env.Config)
		if err != nil {
			return env, fmt.Errorf("setting up tracking worker: %w", err)
		}
		env.TrackingWorker = w
	}

	return env, nil
}
