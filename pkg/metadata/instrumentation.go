package metadata

import (
	"log"
	"strings"

	"github.com/palestamp/barnacle/pkg/api"
)

func WithLogging(next api.MetadataStorage) api.MetadataStorage {
	return &logging{next: next}
}

type logging struct {
	next api.MetadataStorage
}

func (l *logging) RegisterQueueMetadata(rqr api.RegisterQueueRequest) error {
	log.Printf("MetadataStorage.RegisterQueueMetadata [qid=%s]", rqr.QueueID)
	return l.next.RegisterQueueMetadata(rqr)
}

func (l *logging) SetQueueState(qid api.QueueID, state api.QueueState) error {
	log.Printf("MetadataStorage.SetQueueState [qid=%s; state=%s]", qid, state)
	return l.next.SetQueueState(qid, state)
}

func (l *logging) DeleteQueueMetadata(qid api.QueueID) error {
	log.Printf("MetadataStorage.DeleteQueueMetadata [qid=%s]", qid)
	return l.next.DeleteQueueMetadata(qid)
}

func (l *logging) GetQueueMetadata(qid api.QueueID, allowedStates ...api.QueueState) (api.QueueMetadata, error) {
	states := statesSliceToStringSlice(allowedStates)
	log.Printf("MetadataStorage.GetQueueMetadata [qid=%s; als=%s]", qid, strings.Join(states, ", "))
	return l.next.GetQueueMetadata(qid, allowedStates...)
}

func (l *logging) RegisterResource(rm api.ResourceMetadata) error {
	log.Printf("MetadataStorage.RegisterResource [rid=%s]", rm.ResourceID)
	return l.next.RegisterResource(rm)
}
