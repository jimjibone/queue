package queue

import "sync"

// Pub is a channel-based broadcast FIFO queue. Built on top of Queue, the Pub
// sends published messages to all active Subs created via the NewSub method.
type Pub struct {
	nextid int
	subs   map[int]*Sub
	mu     sync.Mutex
}

// Sub is a channel-based FIFO queue, receiving messages from its parent Pub.
type Sub struct {
	parent *Pub
	id     int
	queue  *Queue
}

// NewPub returns a new Pub. Remember to call Close on the Pub once you're
// finished with it.
func NewPub() *Pub {
	return &Pub{
		subs: make(map[int]*Sub),
	}
}

// Close the Pub.
func (p *Pub) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, sub := range p.subs {
		sub.queue.Close()
	}
}

// NewSub adds a new Sub to the Pub and returns it. Remember to
// call Close on the Sub when finished.
func (p *Pub) NewSub() *Sub {
	p.mu.Lock()
	defer p.mu.Unlock()
	sub := &Sub{
		parent: p,
		id:     p.nextid,
		queue:  New(),
	}
	p.nextid++
	p.subs[sub.id] = sub
	return sub
}

// removeSub removes the subscriber from the Pub.
func (p *Pub) removeSub(sub *Sub) {
	p.mu.Lock()
	defer p.mu.Unlock()
	sub.queue.Close()
	if _, found := p.subs[sub.id]; found {
		delete(p.subs, sub.id)
	}
}

// Pub publishes the new item to subscribers.
func (p *Pub) Pub(item interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, sub := range p.subs {
		sub.queue.Push(item)
	}
}

// Sub receives the next published item.
func (s *Sub) Sub() <-chan interface{} {
	return s.queue.Pop()
}

// Close the Sub.
func (s *Sub) Close() {
	s.parent.removeSub(s)
}
