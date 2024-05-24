package source

import (
	"time"
)

type Repository struct {
	Owner, Name string // TODO(adam): defaults to Github
}

type Branch struct {
	Name       string
	Repository Repository
}

type Commit struct {
	Hash       string
	Repository Repository
	Branch     Repository

	Author  string
	Date    time.Time
	Message string
}
