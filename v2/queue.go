package queue

import (
	"sync"
)

// Queue is a channel-based FIFO queue. Similar to a Go channel, items can be
// pushed to the back of the Queue and then popped off the front by listening on
// the Pop channel. This structure differs from channels in that its buffer is
// effectively endless.
type Queue[T any] struct {
	push  chan T
	pop   chan T
	flush chan struct{}
	stop  chan struct{}
	wg    *sync.WaitGroup
}

// New returns a new, running, Queue. Remember to call Close on the Queue once
// you're finished with it.
func New[T any]() *Queue[T] {
	q := &Queue[T]{
		push:  make(chan T),
		pop:   make(chan T),
		flush: make(chan struct{}),
		stop:  make(chan struct{}, 1),
		wg:    &sync.WaitGroup{},
	}
	q.wg.Add(1)
	go q.runloop()
	return q
}

// Close the Queue.
func (q *Queue[T]) Close() {
	select {
	case q.stop <- struct{}{}:
	default:
	}
	q.wg.Wait()
}

// Push an item onto the back of the Queue.
func (q *Queue[T]) Push(item T) {
	q.push <- item
}

// Pop an item from the front of the Queue.
func (q *Queue[T]) Pop() <-chan T {
	return q.pop
}

// Flush empties the Queue.
func (q *Queue[T]) Flush() {
	q.flush <- struct{}{}
}

func (q *Queue[T]) runloop() {
	defer q.wg.Done()
	defer close(q.pop)

	var l []T

	for {
		// Wait for new items to add to the list or stop.
		select {
		case <-q.flush:
		case <-q.stop:
			return
		case item := <-q.push:
			l = append(l, item)
		}

		// While there are items in the list, try to pop them out, otherwise
		// accept new items or stop.
		for len(l) > 0 {
			popItem := l[0]
			select {
			case <-q.flush:
				// Remove all items from the list.
				l = nil
			case <-q.stop:
				return
			case item := <-q.push:
				l = append(l, item)
			case q.pop <- popItem:
				// The item was popped successfully so remove it from the list.
				l = l[1:]
			}
		}
	}
}
