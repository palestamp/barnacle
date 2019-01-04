package api

import (
	"errors"
	"regexp"
)

var (
	// ErrQueueIDInvalid ...
	ErrQueueIDInvalid = errors.New("queue id invalid")
	// ErrResourceIDInvalid ...
	ErrResourceIDInvalid = errors.New("resource id invalid")

	queueIDPattern    = regexp.MustCompile(`^[_a-z][_a-z0-9]*$`)
	resourceIDPattern = queueIDPattern
)

type QueueState string

const (
	// InactiveQueueState is a state in which queue appears until queue
	// resources provisioned, for example postgres table.
	// To provision queue we need it's params to be stored in MetadataStorage.
	InactiveQueueState QueueState = "inactive"

	// ActiveQueueState - queue is active and fully operational
	ActiveQueueState QueueState = "active"
)

// QueueType ...
type QueueType string

const (
	// SimpleDelayQueue ...
	SimpleDelayQueue QueueType = "simple-delay"
)

// QueueID identifier
type QueueID string

func (q *QueueID) Validate() error {
	if !queueIDPattern.MatchString(string(*q)) {
		return ErrQueueIDInvalid

	}
	return nil
}

// ResourceID identifier
type ResourceID string

func (q *ResourceID) Validate() error {
	if !resourceIDPattern.MatchString(string(*q)) {
		return ErrResourceIDInvalid
	}
	return nil
}

// BackendType ...
type BackendType string

type QueueOptions map[string]interface{}

type ResourceConnOptions map[string]interface{}

// QueueMetadata ...
type QueueMetadata struct {
	QueueID     QueueID
	ResourceID  ResourceID
	BackendType BackendType
	QueueType   QueueType
	QueueState  QueueState
	Options     QueueOptions
	ConnOptions ResourceConnOptions
}

type ResourceMetadata struct {
	ResourceID  ResourceID          `json:"id"`
	ConnOptions ResourceConnOptions `json:"options"`
}
