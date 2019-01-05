package postgres

import (
	"github.com/jackc/pgx"

	"github.com/palestamp/barnacle/pkg/api"
	"github.com/palestamp/barnacle/pkg/machinery/decode"
)

type ResourceConnOptions struct {
	URI string `mapstructure:"uri"`
}

func (ops *ResourceConnOptions) verifyKey() string {
	return "uri:" + ops.URI
}

func NewConnector() api.Connector {
	return &connector{
		cache: make(map[api.ResourceID]connectorCacheEntry),
	}
}

type connectorCacheEntry struct {
	backend   api.Backend
	verifyKey string
}

type connector struct {
	cache map[api.ResourceID]connectorCacheEntry
}

func (c *connector) Connect(rid api.ResourceID, ops api.ResourceConnOptions) (api.Backend, error) {
	var op ResourceConnOptions
	if err := decode.Decode(ops, &op); err != nil {
		return nil, err
	}

	if entry := c.lookupCache(rid, op.verifyKey()); entry != nil {
		return entry.backend, nil
	}

	connConfig, err := pgx.ParseURI(op.URI)
	if err != nil {
		return nil, err
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: connConfig,
	})
	if err != nil {
		return nil, err
	}

	backend := NewBackendFromPool(pool)
	c.setCache(rid, op.verifyKey(), backend)
	return backend, nil
}

func (c *connector) lookupCache(rid api.ResourceID, verifyKey string) *connectorCacheEntry {
	entry, ok := c.cache[rid]
	if !ok {
		return nil
	}

	if entry.verifyKey != verifyKey {
		delete(c.cache, rid)
		return nil
	}

	return &entry
}

func (c *connector) setCache(rid api.ResourceID, verifyKey string, backend api.Backend) {
	c.cache[rid] = connectorCacheEntry{
		backend:   backend,
		verifyKey: verifyKey,
	}
}
