package metadata

import (
	"github.com/palestamp/barnacle/pkg/api"
)

type queueEventType int

const (
	queueRegistered = iota + 1
	queueActivated
	queueDeleted
)

type QueueEvent struct {
	QueueID api.QueueID    `json:"qid"`
	Type    queueEventType `json:"type"`
}
