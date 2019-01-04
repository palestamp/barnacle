package api

// MetadataStorage defines behavior for configuration storage.
type MetadataStorage interface {
	RegisterQueueMetadata(RegisterQueueRequest) error
	SetQueueState(QueueID, QueueState) error
	DeleteQueueMetadata(QueueID) error
	GetQueueMetadata(qid QueueID, allowedStates ...QueueState) (QueueMetadata, error)
	RegisterResource(ResourceMetadata) error
}

// Connector ...
type Connector interface {
	Connect(ResourceConnOptions) (Backend, error)
}

// Backend exposes interface for managing queue objects.
type Backend interface {
	// Create queue with QueueMetadata
	CreateQueue(RegisterQueueRequest) error

	// Connect to queue with QueueMetadata
	ConnectToQueue(QueueMetadata) (Queue, error)
}
