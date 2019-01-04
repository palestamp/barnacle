package backends

import (
	"errors"

	"github.com/palestamp/barnacle/pkg/api"
)

var (
	// ErrConnectorNotFound - backend not found
	ErrConnectorNotFound = errors.New("connector not found")
)

// Registry holds known backend types.
type Registry struct {
	connectors map[api.BackendType]api.Connector
}

// NewRegistry creates new Registry instance.
func NewRegistry() *Registry {
	return &Registry{
		connectors: make(map[api.BackendType]api.Connector),
	}
}

// RegisterConnector ...
func (rs *Registry) RegisterConnector(t api.BackendType, connector api.Connector) {
	rs.connectors[t] = connector
}

func (rs *Registry) Connector(t api.BackendType) (api.Connector, error) {
	connector, ok := rs.connectors[t]
	if !ok {
		return nil, ErrConnectorNotFound
	}
	return connector, nil
}
