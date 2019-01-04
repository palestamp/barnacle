package api

import (
	"context"
	"time"
)

// The Queue interface is implemented by objects that
// represent queue
type Queue interface {
	Add(EnqueueMessageRequest) (MessageID, error)
	Ack(ackKey string) error
	Poll(PollRequest) ([]Message, error)
}

// Waiter defines behavior for calculating sleep duration on poll loops
type Waiter interface {
	CalculateSleep(deadlineIn time.Duration) time.Duration
}

// Poll queue.
func Poll(queue Queue, limit int, timeout time.Duration, visibility time.Duration, waiter Waiter) ([]Message, error) {
	deadline := time.Now().Add(timeout)
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	events := make([]Message, 0, limit)
	numToFetch := limit
loop:
	for {
		select {
		case <-timer.C:
			break loop
		default:
			evs, err := queue.Poll(PollRequest{
				Limit:      numToFetch,
				Deadline:   deadline,
				Visibility: visibility,
			})
			if err != nil && err != context.DeadlineExceeded {
				return append(events, evs...), err
			}

			events = append(events, evs...)
			numToFetch -= len(evs)

			if numToFetch <= 0 {
				break loop
			}

			if len(evs) == 0 {
				deadlineIn := deadline.Sub(time.Now())
				time.Sleep(waiter.CalculateSleep(deadlineIn))
			}
		}
	}

	return events, nil
}
