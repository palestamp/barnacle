package metadata

import (
	"encoding/json"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"

	"github.com/palestamp/barnacle/pkg/api"
)

var (
	// ErrQueueNotFound ...
	ErrQueueNotFound = errors.New("queue not found")
)

var _ api.MetadataStorage = (*PostgresMetadataStorage)(nil)

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

func (s *PostgresMetadataStorage) RegisterQueueMetadata(qmi api.RegisterQueueRequest) error {
	b, err := json.Marshal(qmi.Options)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(
		`insert into barnacle.queue_configs (
			queue_id,
			resource_id,
			backend_type,
			queue_type,
			config,
			queue_state
		) values ($1, $2, $3, $4, $5, 'inactive')`,
		qmi.QueueID, qmi.ResourceID, qmi.BackendType, qmi.QueueType, b)
	return errors.Wrap(err, "queue registration failed")
}

func (s *PostgresMetadataStorage) SetQueueState(qid api.QueueID, state api.QueueState) error {
	_, err := s.pool.Exec(
		`update barnacle.queue_configs
			set queue_state = $1
			where queue_id = $2`, string(state), qid)
	return errors.Wrap(err, "queue state change failed")
}

func (s *PostgresMetadataStorage) DeleteQueueMetadata(qid api.QueueID) error {
	_, err := s.pool.Exec(`delete from barnacle.queue_configs where queue_id = $1`, qid)
	return errors.Wrap(err, "queue deletion failed")
}

func (s *PostgresMetadataStorage) GetQueueMetadata(qid api.QueueID, allowedStates ...api.QueueState) (api.QueueMetadata, error) {
	states := statesSliceToStringSlice(allowedStates)

	row := s.pool.QueryRow(`
		select
			qc.queue_id,
			qc.resource_id,
			qc.backend_type,
			qc.queue_type,
			qc.queue_state,
			qc.config,
			rc.config
		from barnacle.queue_configs as qc
		join barnacle.resource_configs as rc using(resource_id)
		where qc.queue_id = $1 and qc.queue_state = ANY($2)`, qid, states)

	var (
		queueID, resourceID, backendType, queueType, queueState string
		queueConfig, resourceConfig                             []byte
	)
	err := row.Scan(
		&queueID,
		&resourceID,
		&backendType,
		&queueType,
		&queueState,
		&queueConfig,
		&resourceConfig)

	if err != nil {
		return api.QueueMetadata{}, ErrQueueNotFound
	}

	var qps api.QueueOptions
	var rps api.ResourceConnOptions

	err = json.Unmarshal(queueConfig, &qps)
	if err != nil {
		return api.QueueMetadata{}, err
	}

	err = json.Unmarshal(resourceConfig, &rps)
	if err != nil {
		return api.QueueMetadata{}, err
	}

	return api.QueueMetadata{
		QueueID:     api.QueueID(queueID),
		ResourceID:  api.ResourceID(resourceID),
		BackendType: api.BackendType(backendType),
		QueueType:   api.QueueType(queueType),
		QueueState:  api.QueueState(queueState),
		Options:     qps,
		ConnOptions: rps,
	}, nil
}

func (s *PostgresMetadataStorage) RegisterResource(rm api.ResourceMetadata) error {
	b, err := json.Marshal(rm.ConnOptions)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(
		`insert into barnacle.resource_configs (resource_id, config) values ($1, $2)`,
		rm.ResourceID, b)
	return errors.Wrap(err, "resource configuration persist call failed")
}

func statesSliceToStringSlice(els []api.QueueState) []string {
	out := make([]string, len(els))
	for i := range els {
		out[i] = string(els[i])
	}
	return out
}
