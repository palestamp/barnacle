package service

import (
	"time"

	"github.com/palestamp/barnacle/pkg/api"
)

type Service struct {
	qms     api.QueueMetadataStorage
	backend api.Backend
}

func New(backend api.Backend, qms api.QueueMetadataStorage) *Service {
	return &Service{qms: qms, backend: backend}
}

func (s *Service) CreateQueue(qm api.QueueMetadata) error {
	if err := s.backend.Create(qm); err != nil {
		return err
	}

	if err := s.qms.RegisterQueue(qm); err != nil {
		return err
	}
	return nil
}

func (s *Service) CreateMessage(msg api.MessageInput) (api.MessageID, error) {
	queue, err := s.connectQueue(msg.ID)
	if err != nil {
		return "", err
	}

	return queue.Add(msg)
}

func (s *Service) AckMessage(qid api.QueueID, ackKey string) error {
	queue, err := s.connectQueue(qid)
	if err != nil {
		return err
	}

	return queue.Ack(ackKey)
}

func (s *Service) PollQueue(id api.QueueID, limit int, timeout, visibility time.Duration) ([]api.Message, error) {
	queue, err := s.connectQueue(id)
	if err != nil {
		return nil, err
	}

	return api.Poll(queue, limit, timeout, visibility, &staticWaiter{})
}

func (s *Service) connectQueue(id api.QueueID) (api.Queue, error) {
	meta, err := s.qms.GetQueueMetadata(id)
	if err != nil {
		return nil, err
	}

	return s.backend.Connect(meta)
}

type staticWaiter struct {
	max time.Duration
}

func (w *staticWaiter) CalculateSleep(deadlineIn time.Duration) time.Duration {
	if w.max > deadlineIn {
		return deadlineIn
	}
	return w.max
}
