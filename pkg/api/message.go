package api

import (
	"encoding/json"
	"time"
)

type MessageID string

type Delay struct {
	time.Duration
}

func (d *Delay) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err == nil {
		duration, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		d.Duration = duration
		return nil
	}

	var n int64
	err = json.Unmarshal(b, &n)
	if err == nil {
		d.Duration = time.Duration(n) * time.Second
		return nil
	}
	return err
}

type MessageInput struct {
	ID    QueueID `json:"topic"`
	Delay Delay   `json:"delay"`
	Data  string  `json:"data"`
}

type Message struct {
	ID          MessageID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Data        string    `json:"data"`
	AckKey      string    `json:"ack_key"`
}

type PollRequest struct {
	Limit      int
	Deadline   time.Time
	Visibility time.Duration
}
