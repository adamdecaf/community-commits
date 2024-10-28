package community_commits

import (
	"log/slog"

	"github.com/adamdecaf/community-commits/internal/forks"
	"github.com/adamdecaf/community-commits/internal/tracker"
)

type Environment struct {
	Logger *slog.Logger
	Config tracker.Config

	ForksRepository forks.Repository
	TrackingWorker  *tracker.Worker
}
