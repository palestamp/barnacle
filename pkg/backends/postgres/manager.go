package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/stretchr/objx"

	"github.com/palestamp/barnacle/pkg/api"
)

type PostgresQueueType string

var (
	ErrUnknownTopicType = errors.New("unknown topic type")

	PSimpleQueue PostgresQueueType = "simple"
)

type backendCreator func(*pgx.ConnPool) (api.Backend, error)

type PostgresQueueProxy struct {
	pool  *pgx.ConnPool
	types map[PostgresQueueType]backendCreator
}

func NewPostgresBackend(uri string) (api.Backend, error) {
	connConfig, err := pgx.ParseURI(uri)
	if err != nil {
		return nil, err
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: connConfig,
		AfterConnect: func(*pgx.Conn) error {
			fmt.Println("connect")
			return nil
		},
	})

	if err != nil {
		return nil, err
	}

	return &PostgresQueueProxy{
		pool: pool,
		types: map[PostgresQueueType]backendCreator{
			PSimpleQueue: NewSimpleQueueManager,
		},
	}, nil
}

func (s *PostgresQueueProxy) Create(metadata api.QueueMetadata) error {
	backend, err := s.getBackend(metadata)
	if err != nil {
		return err
	}

	return backend.Create(metadata)
}

func (s *PostgresQueueProxy) Connect(metadata api.QueueMetadata) (api.Queue, error) {
	backend, err := s.getBackend(metadata)
	if err != nil {
		return nil, err
	}

	return backend.Connect(metadata)
}

func (s *PostgresQueueProxy) getBackend(metadata api.QueueMetadata) (api.Backend, error) {
	obj := objx.Map(metadata.Options)
	topicType := obj.Get("type").Str("simple")

	topicTypeCreator, ok := s.types[PostgresQueueType(topicType)]
	if !ok {
		return nil, ErrUnknownTopicType
	}
	return topicTypeCreator(s.pool)
}
