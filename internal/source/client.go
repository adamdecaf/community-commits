package source

import (
	"context"
)

type Client interface {
	ListNetworkPushEvents(ctx context.Context, repo Repository) ([]PushEvent, error)
}
