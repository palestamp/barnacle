package postgres

import (
	"context"

	"fmt"
	"strconv"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/stretchr/objx"

	"github.com/palestamp/barnacle/pkg/api"
	"github.com/palestamp/barnacle/pkg/backends"
)

var simpleQueueOptionsSchema = `
{
    "$id": "1",
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "SimpleQueueMetadata",
    "oneOf": [
        {
            "type": "object",
            "properties": {
                "table": {
                    "description": "Queue name",
                    "type": "string",
                    "pattern": "^[a-z][a-z0-9_]{0,32}$"
                }
            },
            "additionalProperties": false
        },
        {
            "type": "null"
        }
    ]
}
`

func NewSimpleQueueManager(pool *pgx.ConnPool) (api.Backend, error) {
	validator, err := backends.NewJSONSchemaValidator(gojsonschema.NewStringLoader(simpleQueueOptionsSchema))
	if err != nil {
		return nil, err
	}

	backend := &simpleQueueManager{pool: pool}
	return backends.NewValidationMiddleware(backend, validator), nil
}

type simpleQueueManager struct {
	pool *pgx.ConnPool
}

func defaultTableName(topic api.QueueID) string {
	return string(topic)
}

func (s *simpleQueueManager) Create(metadata api.QueueMetadata) error {
	obj := objx.Map(metadata.Options)
	tableName := obj.Get("table").Str(defaultTableName(metadata.QueueID))

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
	`, tableName, tableName, tableName)

	_, err := s.pool.Exec(stmt)
	return err
}

func (s *simpleQueueManager) Delete(metadata api.QueueMetadata) error {
	obj := objx.Map(metadata.Options)
	tableName := obj.Get("table").Str(defaultTableName(metadata.QueueID))

	stmt := fmt.Sprintf("DROP TABLE %s", tableName)
	_, err := s.pool.Exec(stmt)
	return err
}

func (s *simpleQueueManager) Connect(metadata api.QueueMetadata) (api.Queue, error) {
	obj := objx.Map(metadata.Options)
	tableName := obj.Get("table").Str(defaultTableName(metadata.QueueID))
	return newSimpleDelayQueue(s.pool, tableName)
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
		ack_token = md5(random()::text)
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

func (t *simpleDelayQueue) Add(event api.MessageInput) (api.MessageID, error) {
	delay := int64(event.Delay.Seconds())
	stmt := fmt.Sprintf(`
		INSERT INTO 
		queues.%s(data, scheduled_at, visible_at) VALUES
			($1, NOW() + interval '%d seconds',  NOW() + interval '%d seconds') RETURNING message_id`,
		t.table, delay, delay)

	var eventID int64
	err := t.pool.QueryRow(stmt, event.Data).Scan(&eventID)
	return formatMessageID(eventID), err
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
