package tracker

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/adamdecaf/community-commits/internal/source"
)

// TODO(adam):
//
// Grab repository and forks, traverse each branch collecting commits
// dedup commits, but grab new commits, show if they're ahead of mainline,
// can be merged, how big they are, summary, etc
//
// Besides forks, search for commit hashes in mainline across Github, Gitlab, etc
//
// Web UI with button to cherry-pick commit into PR on the mainline.

type Worker struct {
	conf   Config
	logger *slog.Logger
}

func NewWorker(logger *slog.Logger, conf Config) (*Worker, error) {
	w := &Worker{
		conf:   conf,
		logger: logger,
	}

	return w, nil
}

func (w *Worker) Sync() error {
	// For each repository grab the forks and insert each as an item to crawl
	for _, repo := range w.conf.Tracking.Repositories {
		// TODO(adam):
		fmt.Printf("repo: %#v\n", repo)
	}

	// type Repository
	// 	Source string
	// 	Owner  string
	// 	Name   string

	return nil
}

var (
	sourceClientLock  sync.Mutex
	sourceClientCache = make(map[string]source.Client)
)

func (w *Worker) getSourceClient(name string) source.Client {
	sourceClientLock.Lock()
	defer sourceClientLock.Unlock()

	name = strings.ToLower(name)

	cc, exists := sourceClientCache[name]
	if cc != nil && exists {
		return cc
	}

	cc, err := source.ByName(name, w.conf.Sources)
	if err != nil {
		w.logger.Error(fmt.Sprintf("creating %s source client: %v", name, err))
		return nil
	}
	sourceClientCache[name] = cc

	return cc
}
