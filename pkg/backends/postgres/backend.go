package postgres

import (
	"errors"

	"github.com/jackc/pgx"

	"github.com/palestamp/barnacle/pkg/api"
)

type PostgresQueueType string

var (
	ErrUnknownQueueType = errors.New("unknown queue type")

	queueTypes = map[api.QueueType]managerInitializer{
		api.SimpleDelayQueue: NewDelayQueueManager,
	}
)

type managerInitializer func(*pgx.ConnPool) (api.Manager, error)

type PostgresBackend struct {
	pool *pgx.ConnPool
}

func NewBackendFromPool(pool *pgx.ConnPool) *PostgresBackend {
	return &PostgresBackend{
		pool: pool,
	}
}

func (s *PostgresBackend) GetQueueManager(qt api.QueueType) (api.Manager, error) {
	queueManagerCreator, ok := queueTypes[qt]
	if !ok {
		return nil, ErrUnknownQueueType
	}
	return queueManagerCreator(s.pool)
}
