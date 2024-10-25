package source

import (
	"time"
)

type Repository struct {
	Owner, Name string // TODO(adam): defaults to Github
}

type Branch struct {
	Name string

	LastCommitTimestamp time.Time

	Repository Repository
}

type Commit struct {
	Hash       string
	Repository Repository
	Branch     Branch

	Author  string
	Date    time.Time
	Message string
}
