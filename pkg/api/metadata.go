package api

import (
	"errors"
	"regexp"
)

var ErrQueueNameInvalid = errors.New("queue name invalid")

// QueueID identifier
type QueueID string

// BackendID identifier
type BackendID string
type QueueOptions map[string]interface{}

// OptionsValidator describes behavior to validate QueueOptions
type OptionsValidator interface {
	Validate(QueueOptions) error
}

// Validate delegates QueueOptions validation to OptionsValidator
func (opt QueueOptions) Validate(validator OptionsValidator) error {
	return validator.Validate(opt)
}

// QueueMetadata ...
type QueueMetadata struct {
	QueueID   QueueID      `json:"name"`
	BackendID BackendID    `json:"backend"`
	Options   QueueOptions `json:"options"`
}

var queueNamePattern = regexp.MustCompile(`^[_a-z][_a-z0-9]*$`)

// Validate checks QueueMetadata for misconfiguration.
func (qm *QueueMetadata) Validate() error {
	if !queueNamePattern.MatchString(string(qm.QueueID)) {
		return ErrQueueNameInvalid
	}
	return nil
}
