package source

import (
	"context"
)

type Client interface {
	GetRepository(ctx context.Context, owner, name string) (*Repository, error)
	ListBranches(ctx context.Context, repo Repository) ([]Branch, error)
	ListCommits(ctx context.Context, repo Repository, branch Branch) ([]Commit, error)
}
