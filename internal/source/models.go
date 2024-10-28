package source

import "time"

type Repository struct {
	Owner, Name string // TODO(adam): defaults to Github
}

type PushEvent struct {
	RepoSlug  string
	Commits   []WebCommit
	CreatedAt time.Time
}

type WebCommit struct {
	CommitURL string
	Message   string
}
