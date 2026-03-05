package handlers

import (
	"sync"

	"github.com/craigstjean/vertexmiddleware/config"
	"github.com/craigstjean/vertexmiddleware/vertex"
)

// safeClientCache caches Vertex clients keyed by credential file path,
// so we reuse the same oauth2 TokenSource (and its cached tokens) across requests.
type safeClientCache struct {
	mu      sync.Mutex
	clients map[string]*vertex.Client
}

func newSafeClientCache() *safeClientCache {
	return &safeClientCache{clients: make(map[string]*vertex.Client)}
}

func (cc *safeClientCache) get(kc config.KeyConfig) (*vertex.Client, error) {
	key := kc.CredentialFile + "|" + kc.ProjectID + "|" + kc.Location

	cc.mu.Lock()
	defer cc.mu.Unlock()

	if c, ok := cc.clients[key]; ok {
		return c, nil
	}

	c, err := vertex.NewClient(kc)
	if err != nil {
		return nil, err
	}
	cc.clients[key] = c
	return c, nil
}
