package postgres

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"

	"github.com/palestamp/barnacle/pkg/api"
	"github.com/palestamp/barnacle/pkg/machinery/decode"
)

var ErrTableNameInvalid = errors.New("table name invalid")

func NewDelayQueueManager(pool *pgx.ConnPool) (api.Backend, error) {
	return &delayQueueManager{pool: pool}, nil
}

type delayQueueManager struct {
	pool *pgx.ConnPool
}

var queueTableNamePattern = regexp.MustCompile(`[a-z][a-z0-9_]{0,31s}`)

type delayQueueOptions struct {
	Table string `mapstructure:"table"`
}

func (dq *delayQueueOptions) Validate() error {
	if !queueTableNamePattern.MatchString(dq.Table) {
		return ErrTableNameInvalid
	}
	return nil
}

func (s *delayQueueManager) decodeOpts(qm api.QueueOptions) (delayQueueOptions, error) {
	var ops delayQueueOptions
	if err := decode.Decode(qm, &ops); err != nil {
		return ops, err
	}

	err := ops.Validate()
	return ops, err
}

func (s *delayQueueManager) CreateQueue(rqr api.RegisterQueueRequest) error {
	ops, err := s.decodeOpts(rqr.Options)
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf(`
	CREATE TABLE queues.%s (
		message_id BIGSERIAL PRIMARY KEY,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
		visible_at TIMESTAMP WITH TIME ZONE NOT NULL,
		ack_token varchar(32),
		attempts int NOT NULL DEFAULT 0,
		data text
	);
	CREATE INDEX idx_%s_visible_at ON queues.%s (visible_at);
	`, ops.Table, ops.Table, ops.Table)

	_, err = s.pool.Exec(stmt)
	return err
}

func (s *delayQueueManager) Delete(qm api.QueueMetadata) error {
	ops, err := s.decodeOpts(qm.Options)
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf("DROP TABLE %s", ops.Table)
	_, err = s.pool.Exec(stmt)
	return err
}

func (s *delayQueueManager) ConnectToQueue(qm api.QueueMetadata) (api.Queue, error) {
	ops, err := s.decodeOpts(qm.Options)
	if err != nil {
		return nil, err
	}

	return newSimpleDelayQueue(s.pool, ops.Table)
}

type simpleDelayQueue struct {
	pool  *pgx.ConnPool
	table string
}

func newSimpleDelayQueue(pool *pgx.ConnPool, tableName string) (*simpleDelayQueue, error) {
	tp := &simpleDelayQueue{
		pool:  pool,
		table: tableName,
	}
	return tp, nil
}

func (t *simpleDelayQueue) Poll(pr api.PollRequest) ([]api.Message, error) {
	stmt := fmt.Sprintf(`
	UPDATE queues.%s as original
	SET 
		visible_at = NOW() + interval '%d seconds',
		attempts = attempts + 1,
		ack_token = substring(md5(random()::text) from 1 for 7)
	FROM (
		SELECT
			message_id
		FROM
			queues.%s
		WHERE visible_at <= NOW()
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	) as subquery
	WHERE original.message_id = subquery.message_id
	RETURNING
		original.message_id,
		original.created_at,
		original.scheduled_at,
		original.data,
		original.ack_token
	`, t.table, int64(pr.Visibility.Seconds()), t.table)

	ctx, cancel := context.WithDeadline(context.Background(), pr.Deadline)
	defer cancel()

	rows, err := t.pool.QueryEx(ctx, stmt, nil, pr.Limit)
	if err != nil {
		return nil, err
	}

	out := make([]api.Message, 0, pr.Limit)
	for rows.Next() {
		var messageID int64
		var ackToken string
		var message api.Message
		if err := rows.Scan(
			&messageID,
			&message.CreatedAt,
			&message.ScheduledAt,
			&message.Data,
			&ackToken,
		); err != nil {
			return nil, err
		}

		message.ID = formatMessageID(messageID)
		message.AckKey = formatAckKey(messageID, ackToken)
		out = append(out, message)
	}

	return out, nil
}

func (t *simpleDelayQueue) Add(emr api.EnqueueMessageRequest) (api.MessageID, error) {
	delay := int64(emr.Delay.Seconds())
	stmt := fmt.Sprintf(`
		INSERT INTO 
		queues.%s(data, scheduled_at, visible_at) VALUES
			($1, NOW() + interval '%d seconds',  NOW() + interval '%d seconds') RETURNING message_id`,
		t.table, delay, delay)

	var messageID int64
	err := t.pool.QueryRow(stmt, emr.Data).Scan(&messageID)
	return formatMessageID(messageID), err
}

func (t *simpleDelayQueue) Ack(ackKey string) error {
	id, token, err := parseAckKey(ackKey)
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf(`DELETE FROM queues.%s WHERE message_id = $1 AND ack_token = $2;`, t.table)
	ct, err := t.pool.Exec(stmt, id, token)
	if err != nil {
		return err
	}

	if ct.RowsAffected() <= 0 {
		return errors.New("ack ineffective")
	}
	return nil
}

func parseAckKey(s string) (int64, string, error) {
	toks := strings.Split(s, "/")
	if len(toks) != 2 {
		return 0, "", errors.New("invalid ack key")
	}

	id, err := strconv.ParseInt(toks[0], 10, 64)
	if err != nil {
		return 0, "", errors.Wrap(err, "invalid ack key")
	}

	return id, toks[1], nil
}

func formatAckKey(id int64, token string) string {
	return fmt.Sprintf("%d/%s", id, token)
}

func formatMessageID(id int64) api.MessageID {
	return api.MessageID(strconv.FormatInt(id, 10))
}
