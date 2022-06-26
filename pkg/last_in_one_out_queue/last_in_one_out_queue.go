package last_in_one_out_queue

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const pollPeriod = 250 * time.Millisecond

type LastInOneOutQueue[T any] struct {
	queuedItem *T
	ch         chan T
	mutex      sync.Mutex
}

func New[T any]() *LastInOneOutQueue[T] {
	return &LastInOneOutQueue[T]{
		queuedItem: nil,
		ch:         make(chan T),
		mutex:      sync.Mutex{},
	}
}

func (q *LastInOneOutQueue[T]) Run(ctx context.Context) {
	log.Trace().Msg("last-in-one-out queue starting")
	defer log.Trace().Msg("last-in-one-out queue stopping")

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollPeriod):
			// Periodically try pushing the queued item into the channel.
			q.tryPushingItem()
		}
	}
}

// Enqueue puts the item in the queue. It replaces any previously-queued item.
func (q *LastInOneOutQueue[T]) Enqueue(item T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.queuedItem = &item
	q.tryPushingItem_unsafe()
}

// Item returns the channel on which queued items can be dequeued.
func (q *LastInOneOutQueue[T]) Item() <-chan T {
	return q.ch
}

func (q *LastInOneOutQueue[T]) tryPushingItem() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.queuedItem == nil {
		return
	}

	q.tryPushingItem_unsafe()
}

// tryPushingItem_unsafe tries to push the queued item.
// It assumes that q.queuedItem is not nil, and doesn't obtain the mutex.
func (q *LastInOneOutQueue[T]) tryPushingItem_unsafe() {
	// It's fine if pushing to the channel fails; this means that the receiving
	// end of the queue isn't ready to process another item yet.
	select {
	case q.ch <- *q.queuedItem:
		q.queuedItem = nil
		log.Trace().Msg("last-in-one-out queue: queued item is being handled")
	default:
	}
}
