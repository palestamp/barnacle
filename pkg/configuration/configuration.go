package configuration

import (
	"encoding/json"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"

	"github.com/palestamp/barnacle/pkg/api"
)

type PostgresMetadataStorage struct {
	pool *pgx.ConnPool
}

func NewPostgresStorage(uri string) (*PostgresMetadataStorage, error) {
	connConfig, err := pgx.ParseURI(uri)
	if err != nil {
		return nil, err
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: connConfig,
	})

	if err != nil {
		return nil, err
	}
	return &PostgresMetadataStorage{pool: pool}, nil
}

func (s *PostgresMetadataStorage) RegisterQueue(meta api.QueueMetadata) error {
	if err := meta.Validate(); err != nil {
		return err
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(`insert into barnacle.queue_configs (queue_name, config) values ($1, $2)`, meta.QueueID, b)
	return errors.Wrap(err, "queue configuration persist call failed")
}

func (s *PostgresMetadataStorage) GetQueueMetadata(ident api.QueueID) (api.QueueMetadata, error) {
	row := s.pool.QueryRow(`select config from barnacle.queue_configs where queue_name = $1`, ident)
	var c []byte
	if err := row.Scan(&c); err != nil {
		return api.QueueMetadata{}, err
	}

	var qm api.QueueMetadata
	err := json.Unmarshal(c, &qm)
	return qm, err
}
