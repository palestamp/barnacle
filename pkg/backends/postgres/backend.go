package postgres

import (
	"errors"

	"github.com/jackc/pgx"

	"github.com/palestamp/barnacle/pkg/api"
)

type PostgresQueueType string

var (
	ErrUnknownQueueType = errors.New("unknown queue type")

	queueTypes = map[api.QueueType]backendCreator{
		api.SimpleDelayQueue: NewDelayQueueManager,
	}
)

type backendCreator func(*pgx.ConnPool) (api.Backend, error)

type PostgresBackend struct {
	pool *pgx.ConnPool
}

func NewBackendFromPool(pool *pgx.ConnPool) *PostgresBackend {
	return &PostgresBackend{
		pool: pool,
	}
}

func (s *PostgresBackend) CreateQueue(qmi api.RegisterQueueRequest) error {
	m, err := s.getQueueManager(qmi.QueueType)
	if err != nil {
		return err
	}

	return m.CreateQueue(qmi)
}

func (s *PostgresBackend) ConnectToQueue(qm api.QueueMetadata) (api.Queue, error) {
	m, err := s.getQueueManager(qm.QueueType)
	if err != nil {
		return nil, err
	}

	return m.ConnectToQueue(qm)
}

func (s *PostgresBackend) getQueueManager(qt api.QueueType) (api.Backend, error) {
	queueManagerCreator, ok := queueTypes[qt]
	if !ok {
		return nil, ErrUnknownQueueType
	}
	return queueManagerCreator(s.pool)
}
