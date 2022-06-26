package last_in_one_out_queue

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	q := New[string]()
	select {
	case <-q.Item():
		t.Fatal("a new queue shouldn't hold an item")
	default:
	}
}

func TestQueueAndGet(t *testing.T) {
	q := New[string]()
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		q.Enqueue("hey")
	}()

	select {
	case item := <-q.Item():
		assert.Equal(t, "hey", item)
	case <-time.After(200 * time.Millisecond):
		t.Error("enqueueing while waiting for an item should push it immediately")
	}

	wg.Wait()
}

func TestQueueMultiple(t *testing.T) {
	q := New[string]()

	q.Enqueue("hey")
	q.Enqueue("these are multiple items")
	q.Enqueue("the last one should be the")
	q.Enqueue("winner")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		q.Run(ctx)
	}()

	select {
	case item := <-q.Item():
		func() {
			q.mutex.Lock()
			defer q.mutex.Unlock()
			assert.Nil(t, q.queuedItem,
				"after popping an item of the queue, the queue should be empty")
		}()
		assert.Equal(t, "winner", item)

	case <-time.After(10 * pollPeriod):
		t.Error("timeout waiting for item")
	}

	cancel()
	wg.Wait()
}
