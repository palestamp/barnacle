package backends

import (
	"errors"

	"github.com/palestamp/barnacle/pkg/api"
)

var (
	// ErrUnknownBackend - backend not found
	ErrUnknownBackend = errors.New("unknown backend")
)

// Proxy forwards calls to actual implementation based on api.BackendID
type Proxy struct {
	registry map[api.BackendID]api.Backend
}

// NewProxy creates new Proxy instance.
func NewProxy() *Proxy {
	return &Proxy{
		registry: make(map[api.BackendID]api.Backend),
	}
}

// RegisterBackend in Proxy.
func (rs *Proxy) RegisterBackend(id api.BackendID, backend api.Backend) {
	rs.registry[id] = backend
}

// Create forwards Create call to actual backend instance.
func (rs *Proxy) Create(meta api.QueueMetadata) error {
	b, ok := rs.registry[meta.BackendID]
	if !ok {
		return ErrUnknownBackend
	}

	return b.Create(meta)
}

// Connect forwards Connect call to actual backend instance.s
func (rs *Proxy) Connect(meta api.QueueMetadata) (api.Queue, error) {
	b, ok := rs.registry[meta.BackendID]
	if !ok {
		return nil, ErrUnknownBackend
	}

	return b.Connect(meta)
}
