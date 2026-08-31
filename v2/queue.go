package queue

import (
	"context"
	"sync"
)

// Queue is a channel-based FIFO queue. Similar to a Go channel, items can be
// pushed to the back of the Queue and then popped off the front by listening on
// the Pop channel. This structure differs from channels in that its buffer is
// effectively endless.
type Queue[T any] struct {
	push    chan T
	pop     chan T
	flush   chan struct{}
	cancel  func()
	done    <-chan struct{}
	discard chan bool
	wg      *sync.WaitGroup
}

// New returns a new, running, Queue. Remember to call Close on the Queue once
// you're finished with it.
func New[T any]() *Queue[T] {
	ctx, cancel := context.WithCancel(context.Background())
	q := &Queue[T]{
		push:    make(chan T),
		pop:     make(chan T),
		flush:   make(chan struct{}),
		cancel:  cancel,
		done:    ctx.Done(),
		discard: make(chan bool),
		wg:      &sync.WaitGroup{},
	}
	q.wg.Add(1)
	go q.runloop(ctx)
	return q
}

// Close the Queue.
func (q *Queue[T]) Close() {
	q.cancel()
	q.wg.Wait()
}

// Push an item onto the back of the Queue. If the Queue is closed, the item is
// dropped.
func (q *Queue[T]) Push(item T) {
	select {
	case q.push <- item:
	case <-q.done:
	}
}

// Pop an item from the front of the Queue.
func (q *Queue[T]) Pop() <-chan T {
	return q.pop
}

// Flush empties the Queue. A no-op on a closed Queue, per Push.
func (q *Queue[T]) Flush() {
	select {
	case q.flush <- struct{}{}:
	case <-q.done:
	}
}

// Discard all pushed items. A no-op on a closed Queue, per Push.
func (q *Queue[T]) Discard(discard bool) {
	select {
	case q.discard <- discard:
	case <-q.done:
	}
}

func (q *Queue[T]) runloop(ctx context.Context) {
	defer q.wg.Done()
	defer close(q.pop)

	var l []T
	var discard bool

	for {
		// Wait for new items to add to the list or stop.
		select {
		case <-ctx.Done():
			return
		case <-q.flush:
		case discard = <-q.discard:
		case item := <-q.push:
			if !discard {
				l = append(l, item)
			}
		}

		// While there are items in the list, try to pop them out, otherwise
		// accept new items or stop.
		for len(l) > 0 {
			popItem := l[0]
			select {
			case <-ctx.Done():
				return
			case <-q.flush:
				// Remove all items from the list.
				l = nil
			case discard = <-q.discard:
			case item := <-q.push:
				if !discard {
					l = append(l, item)
				}
			case q.pop <- popItem:
				// The item was popped successfully so remove it from the list.
				l = l[1:]
			}
		}
	}
}
