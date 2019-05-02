package metadata

import (
	"encoding/json"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"

	"github.com/palestamp/barnacle/pkg/api"
	"github.com/palestamp/barnacle/pkg/machinery/notify"
)

var (
	// ErrQueueNotFound ...
	ErrQueueNotFound = errors.New("queue not found")
)

var _ api.MetadataStorage = (*PostgresMetadataStorage)(nil)

type PostgresMetadataStorage struct {
	pool               *pgx.ConnPool
	queueMetadataCache map[api.QueueID]api.QueueMetadata
	notifier           *notify.PgNotifier
	listener           *notify.PgListener
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

	s := &PostgresMetadataStorage{
		pool:               pool,
		queueMetadataCache: make(map[api.QueueID]api.QueueMetadata),
		notifier:           notify.NewPgNotifier(pool, "_bcl_mq_events"),
	}

	listener := notify.NewPgListener(pool, "_bcl_mq_events")
	if err := listener.Listen(); err != nil {
		return nil, err
	}

	listener.Register(s)
	s.listener = listener

	return s, nil
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
	if err != nil {
		return err
	}
	s.send(QueueEvent{QueueID: qmi.QueueID, Type: queueRegistered})
	return errors.Wrap(err, "queue registration failed")
}

func (s *PostgresMetadataStorage) SetQueueState(qid api.QueueID, state api.QueueState) error {
	ct, err := s.pool.Exec(
		`update barnacle.queue_configs
			set queue_state = $1
			where queue_id = $2`, string(state), qid)

	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return ErrQueueNotFound
	}

	switch state {
	case api.ActiveQueueState:
		s.send(QueueEvent{QueueID: qid, Type: queueActivated})
	}
	return errors.Wrap(err, "queue state change failed")
}

func (s *PostgresMetadataStorage) DeleteQueueMetadata(qid api.QueueID) error {
	_, err := s.pool.Exec(`delete from barnacle.queue_configs where queue_id = $1`, qid)
	if err != nil {
		return err
	}

	s.send(QueueEvent{QueueID: qid, Type: queueDeleted})
	return errors.Wrap(err, "queue deletion failed")
}

func (s *PostgresMetadataStorage) GetQueueMetadata(qid api.QueueID, allowedStates ...api.QueueState) (api.QueueMetadata, error) {
	if len(allowedStates) == 1 && allowedStates[0] == api.ActiveQueueState {
		if qm, ok := s.lookupCache(qid); ok {
			return qm, nil
		}
	}

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

	nqm := api.QueueMetadata{
		QueueID:     api.QueueID(queueID),
		ResourceID:  api.ResourceID(resourceID),
		BackendType: api.BackendType(backendType),
		QueueType:   api.QueueType(queueType),
		QueueState:  api.QueueState(queueState),
		Options:     qps,
		ConnOptions: rps,
	}

	s.setCache(qid, nqm)
	return nqm, nil
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

func (s *PostgresMetadataStorage) send(event QueueEvent) {
	b, _ := json.Marshal(event)
	s.notifier.Send(b)
}

func (s *PostgresMetadataStorage) Notify(payload []byte) {
	var qev QueueEvent
	if err := json.Unmarshal(payload, &qev); err != nil {
		return
	}

	switch qev.Type {
	case queueActivated:
	case queueRegistered:
	case queueDeleted:
		s.invalidateCacheEntry(qev.QueueID)
	}
}

func (s *PostgresMetadataStorage) lookupCache(qid api.QueueID) (api.QueueMetadata, bool) {
	qm, ok := s.queueMetadataCache[qid]
	return qm, ok
}

func (s *PostgresMetadataStorage) setCache(qid api.QueueID, qm api.QueueMetadata) {
	s.queueMetadataCache[qid] = qm
}

func (s *PostgresMetadataStorage) invalidateCacheEntry(qid api.QueueID) {
	delete(s.queueMetadataCache, qid)
}

func statesSliceToStringSlice(els []api.QueueState) []string {
	out := make([]string, len(els))
	for i := range els {
		out[i] = string(els[i])
	}
	return out
}
