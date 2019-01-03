package api

// QueueMetadataStorage defines behavior for configuration storage.
type QueueMetadataStorage interface {
	RegisterQueue(QueueMetadata) error
	GetQueueMetadata(QueueID) (QueueMetadata, error)
}

// Backend exposes interface for managing queue objects.
type Backend interface {
	// Create queue with QueueMetadata
	Create(QueueMetadata) error

	// Connect to queue with QueueMetadata
	Connect(QueueMetadata) (Queue, error)
}
