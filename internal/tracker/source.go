package tracker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/adamdecaf/community-commits/internal/source"
)

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
