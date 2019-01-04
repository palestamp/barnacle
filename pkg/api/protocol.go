package api

type RegisterQueueRequest struct {
	QueueID     QueueID      `json:"id"`
	ResourceID  ResourceID   `json:"resource"`
	BackendType BackendType  `json:"backend"`
	QueueType   QueueType    `json:"type"`
	Options     QueueOptions `json:"options"`
}

func (r *RegisterQueueRequest) Validate() error {
	return Check(
		Ce(r.QueueID.Validate()),
		Ce(r.ResourceID.Validate()),
		Cb(r.BackendType != "", "backend type can not be empty"),
		Cb(r.QueueType != "", "queue type can not be empty"),
	)
}

type EnqueueMessageRequest struct {
	QueueID QueueID `json:"queue"`
	Delay   Delay   `json:"delay"`
	Data    string  `json:"data"`
}
