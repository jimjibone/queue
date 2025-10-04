package queue

import (
	"container/list"
	"sync"
)

// Queue is a channel-based FIFO queue. Similar to a Go channel, items can be
// pushed to the back of the Queue and then popped off the front by listening on
// the Pop channel. This structure differs from channels in that its buffer is
// effectively endless.
type Queue struct {
	push  chan interface{}
	pop   chan interface{}
	flush chan struct{}
	stop  chan struct{}
	wg    *sync.WaitGroup
}

// New returns a new, running, Queue. Remember to call Close on the Queue once
// you're finished with it.
func New() *Queue {
	q := &Queue{
		push:  make(chan interface{}),
		pop:   make(chan interface{}),
		flush: make(chan struct{}),
		stop:  make(chan struct{}, 1),
		wg:    &sync.WaitGroup{},
	}
	q.wg.Add(1)
	go q.runloop()
	return q
}

// Close the Queue.
func (q *Queue) Close() {
	select {
	case q.stop <- struct{}{}:
	default:
	}
	q.wg.Wait()
}

// Push an item onto the back of the Queue.
func (q *Queue) Push(item interface{}) {
	q.push <- item
}

// Pop an item from the front of the Queue.
func (q *Queue) Pop() <-chan interface{} {
	return q.pop
}

// Flush empties the Queue.
func (q *Queue) Flush() {
	q.flush <- struct{}{}
}

func (q *Queue) runloop() {
	defer q.wg.Done()
	defer close(q.pop)

	l := list.New()

	for {
		// Wait for new items to add to the list or stop.
		select {
		case <-q.flush:
		case <-q.stop:
			return
		case item := <-q.push:
			l.PushBack(item)
		}

		// While there are items in the list, try to pop them out, otherwise
		// accept new items or stop.
		for l.Len() > 0 {
			popItem := l.Front()
			select {
			case <-q.flush:
				// Remove all items from the list.
				for l.Len() > 0 {
					e := l.Front()
					l.Remove(e)
				}
			case <-q.stop:
				return
			case item := <-q.push:
				l.PushBack(item)
			case q.pop <- popItem.Value:
				// The item was popped successfully so remove it from the list.
				l.Remove(popItem)
			}
		}
	}
}
