package forks

import (
	"context"
	"database/sql"

	"github.com/adamdecaf/community-commits/internal/source"
)

type Repository interface {
	SaveCommits(ctx context.Context, repo source.Repository, branch source.Branch, commits []source.Commit) error
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

type sqlRepository struct {
	db *sql.DB
}

func (r *sqlRepository) SaveCommits(ctx context.Context, repo source.Repository, branch source.Branch, commits []source.Commit) error {
	return nil
}
