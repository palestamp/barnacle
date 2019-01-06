package api

// The Queue interface is implemented by objects that
// represent queue
type Queue interface {
	Add(EnqueueMessageRequest) (MessageID, error)
	Ack(ackKey string) error
	Poll(PollRequest) ([]Message, error)
}

// MetadataStorage defines behavior for configuration storage.
type MetadataStorage interface {
	RegisterQueueMetadata(RegisterQueueRequest) error
	SetQueueState(QueueID, QueueState) error
	DeleteQueueMetadata(QueueID) error
	GetQueueMetadata(qid QueueID, allowedStates ...QueueState) (QueueMetadata, error)
	RegisterResource(ResourceMetadata) error
}

// Connector is a factory for Backend creation.
type Connector interface {
	Connect(ResourceID, ResourceConnOptions) (Backend, error)
}

type Manager interface {
	// Create queue with QueueMetadata
	CreateQueue(RegisterQueueRequest) error

	// Connect to queue with QueueMetadata
	ConnectToQueue(QueueMetadata) (Queue, error)
}

// Backend exposes interface for managing queue objects.
type Backend interface {
	// Create queue with QueueMetadata
	GetQueueManager(QueueType) (Manager, error)
}
