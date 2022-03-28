package queue

import "sync"

// Pub is a channel-based broadcast FIFO queue. Built on top of Queue, the Pub
// sends published messages to all active Subs created via the NewSub method.
type Pub[T any] struct {
	nextid int
	subs   map[int]*Sub[T]
	mu     sync.Mutex
}

// Sub is a channel-based FIFO queue, receiving messages from its parent Pub.
type Sub[T any] struct {
	parent *Pub[T]
	id     int
	queue  *Queue[T]
}

// NewPub returns a new Pub. Remember to call Close on the Pub once you're
// finished with it.
func NewPub[T any]() *Pub[T] {
	return &Pub[T]{
		subs: make(map[int]*Sub[T]),
	}
}

// Close the Pub.
func (p *Pub[T]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, sub := range p.subs {
		sub.queue.Close()
	}
}

// NewSub adds a new Sub to the Pub and returns it. Remember to
// call Close on the Sub when finished.
func (p *Pub[T]) NewSub() *Sub[T] {
	p.mu.Lock()
	defer p.mu.Unlock()
	sub := &Sub[T]{
		parent: p,
		id:     p.nextid,
		queue:  New[T](),
	}
	p.nextid++
	p.subs[sub.id] = sub
	return sub
}

// removeSub removes the subscriber from the Pub.
func (p *Pub[T]) removeSub(sub *Sub[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()
	sub.queue.Close()
	if _, found := p.subs[sub.id]; found {
		delete(p.subs, sub.id)
	}
}

// Pub publishes the new item to subscribers.
func (p *Pub[T]) Pub(item T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, sub := range p.subs {
		sub.queue.Push(item)
	}
}

// Sub receives the next published item.
func (s *Sub[T]) Sub() <-chan T {
	return s.queue.Pop()
}

// Close the Sub.
func (s *Sub[T]) Close() {
	s.parent.removeSub(s)
}
