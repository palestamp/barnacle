package apis

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/palestamp/barnacle/pkg/api"
)

type V1APIService interface {
	CreateQueue(api.RegisterQueueRequest) error
	CreateMessage(api.EnqueueMessageRequest) (api.MessageID, error)
	AckMessage(api.QueueID, string) error
	PollQueue(id api.QueueID, limit int, timeout, visibility time.Duration) ([]api.Message, error)
	CreateResource(api.ResourceMetadata) error
}

func NewV1API(svc V1APIService) http.Handler {
	s := &v1API{svc: svc}
	mux := http.NewServeMux()
	mux.Handle("/v1/queues.create", http.HandlerFunc(s.CreateQueue))
	mux.Handle("/v1/messages.create", http.HandlerFunc(s.CreateMessage))
	mux.Handle("/v1/messages.poll", http.HandlerFunc(s.PollMessages))
	mux.Handle("/v1/messages.ack", http.HandlerFunc(s.AckMessage))
	mux.Handle("/v1/resources.create", http.HandlerFunc(s.CreateResource))
	return mux
}

type v1API struct {
	svc V1APIService
}

func (s *v1API) CreateQueue(w http.ResponseWriter, r *http.Request) {
	var m api.RegisterQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), 400)
	}
	defer r.Body.Close()

	if err := s.svc.CreateQueue(m); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (s *v1API) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var emr api.EnqueueMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&emr); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer r.Body.Close()

	messageID, err := s.svc.CreateMessage(emr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(struct {
		ID api.MessageID `json:"id"`
	}{
		ID: messageID,
	})
}

func (s *v1API) AckMessage(w http.ResponseWriter, r *http.Request) {
	qp := r.URL.Query()

	ackKey := qp.Get("key")
	if ackKey == "" {
		http.Error(w, "ackKey must be set", 422)
		return
	}

	queue := qp.Get("queue")
	if queue == "" {
		http.Error(w, "queue must be set", 422)
		return
	}

	err := s.svc.AckMessage(api.QueueID(queue), ackKey)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (s *v1API) PollMessages(w http.ResponseWriter, r *http.Request) {
	qp := r.URL.Query()

	queue := qp.Get("queue")
	if queue == "" {
		http.Error(w, "queue must be set", 422)
		return
	}

	limit := qp.Get("limit")
	if limit == "" {
		http.Error(w, "limit must be set", 422)
		return
	}

	l, err := strconv.Atoi(limit)
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}

	timeout := parseSeconds(qp.Get("timeout"), time.Second)
	visibility := parseSeconds(qp.Get("visibility"), time.Minute)

	mgs, err := s.svc.PollQueue(api.QueueID(queue), l, timeout, visibility)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(struct {
		Messages []api.Message `json:"messages"`
	}{
		Messages: mgs,
	})
}

func (s *v1API) CreateResource(w http.ResponseWriter, r *http.Request) {
	var m api.ResourceMetadata
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), 400)
	}
	defer r.Body.Close()

	if err := s.svc.CreateResource(m); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func parseSeconds(s string, def time.Duration) time.Duration {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return time.Second * time.Duration(v)
}
