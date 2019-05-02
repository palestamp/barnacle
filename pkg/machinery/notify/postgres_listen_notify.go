package notify

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx"
)

type Observer interface {
	Notify([]byte)
}

type PgListener struct {
	pool      *pgx.ConnPool
	channel   string
	observers []Observer
}

func NewPgListener(pool *pgx.ConnPool, channel string) *PgListener {
	return &PgListener{
		pool:      pool,
		channel:   channel,
		observers: make([]Observer, 0),
	}
}

func (l *PgListener) Register(o Observer) {
	l.observers = append(l.observers, o)
}

func (l *PgListener) notifyObservers(ev []byte) {
	for _, o := range l.observers {
		o.Notify(ev)
	}
}

func (l *PgListener) Listen() error {
	conn, err := l.pool.Acquire()
	if err != nil {
		return err
	}

	err = conn.Listen(l.channel)
	if err != nil {
		l.pool.Release(conn)
		return err
	}

	go func() {
		defer l.pool.Release(conn)

		for {
			notification, err := conn.WaitForNotification(context.Background())
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error waiting for notification:", err)
				continue
			}

			if notification == nil {
				fmt.Fprintln(os.Stderr, "nil notification")
				continue
			}

			fmt.Println("PID:", notification.PID, "Channel:", notification.Channel, "Payload:", notification.Payload)
			l.notifyObservers([]byte(notification.Payload))
		}
	}()
	return nil
}

type PgNotifier struct {
	channel string
	pool    *pgx.ConnPool
}

func NewPgNotifier(pool *pgx.ConnPool, channel string) *PgNotifier {
	return &PgNotifier{
		pool:    pool,
		channel: channel,
	}
}

func (n *PgNotifier) Send(payload []byte) error {
	_, err := n.pool.Exec(`SELECT pg_notify($1, $2)`, n.channel, payload)
	return err
}
