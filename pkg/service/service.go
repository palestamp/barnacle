package service

import (
	"errors"
	"time"

	"github.com/palestamp/barnacle/pkg/api"
)

type ConnectorFactory interface {
	Connector(api.BackendType) (api.Connector, error)
}

type Service struct {
	qms              api.MetadataStorage
	connectorFactory ConnectorFactory
}

func New(factory ConnectorFactory, qms api.MetadataStorage) *Service {
	return &Service{qms: qms, connectorFactory: factory}
}

func (s *Service) CreateQueue(qmi api.RegisterQueueRequest) error {
	if err := qmi.Validate(); err != nil {
		return err
	}

	if err := s.qms.RegisterQueueMetadata(qmi); err != nil {
		return err
	}

	if err := s.createQueue(qmi); err != nil {
		err1 := s.qms.DeleteQueueMetadata(qmi.QueueID)
		if err1 != nil {
			return errors.New("fatal: queue registration failed, stale artifacts")
		}
		return err
	}

	return s.qms.SetQueueState(qmi.QueueID, api.ActiveQueueState)
}

func (s *Service) CreateMessage(emr api.EnqueueMessageRequest) (api.MessageID, error) {
	queue, err := s.connectQueueByID(emr.QueueID)
	if err != nil {
		return "", err
	}

	return queue.Add(emr)
}

func (s *Service) AckMessage(qid api.QueueID, ackKey string) error {
	queue, err := s.connectQueueByID(qid)
	if err != nil {
		return err
	}

	return queue.Ack(ackKey)
}

func (s *Service) PollQueue(qid api.QueueID, limit int, timeout, visibility time.Duration) ([]api.Message, error) {
	queue, err := s.connectQueueByID(qid)
	if err != nil {
		return nil, err
	}

	return api.Poll(queue, limit, timeout, visibility, &staticWaiter{})
}

func (s *Service) CreateResource(rm api.ResourceMetadata) error {
	return s.qms.RegisterResource(rm)
}

func (s *Service) connectQueueByID(id api.QueueID) (api.Queue, error) {
	qm, err := s.qms.GetQueueMetadata(id, api.ActiveQueueState)
	if err != nil {
		return nil, err
	}

	backend, err := s.connectBackendByMetadata(qm)
	if err != nil {
		return nil, err
	}

	return backend.ConnectToQueue(qm)
}

func (s *Service) connectBackendByMetadata(qm api.QueueMetadata) (api.Backend, error) {
	connector, err := s.connectorFactory.Connector(qm.BackendType)
	if err != nil {
		return nil, err
	}

	return connector.Connect(qm.ConnOptions)
}

func (s *Service) connectBackend(id api.QueueID, qss ...api.QueueState) (api.Backend, error) {
	qm, err := s.qms.GetQueueMetadata(id, qss...)
	if err != nil {
		return nil, err
	}

	return s.connectBackendByMetadata(qm)
}

func (s *Service) createQueue(qmi api.RegisterQueueRequest) error {
	backend, err := s.connectBackend(qmi.QueueID, api.ActiveQueueState, api.InactiveQueueState)
	if err != nil {
		return err
	}

	return backend.CreateQueue(qmi)
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
