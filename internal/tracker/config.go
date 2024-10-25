package tracker

type Config struct {
	Repositories []Repository
}

type Repository struct {
	Source string
	Owner  string
	Name   string
}
