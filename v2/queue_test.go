package queue_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jimjibone/queue/v2"
)

func TestQueueSimple(t *testing.T) {
	q := queue.New[string]()
	defer q.Close()

	item := "item"
	q.Push(item)
	out := <-q.Pop()
	if out != item {
		t.Errorf("Queue output should be %q but is %q", item, out)
	}

	select {
	case <-q.Pop():
		t.Error("Queue.Pop returned value")
	default:
	}

	// Test closing while there are still items in the Queue.
	q.Push(item)
}

func TestQueueOrdered(t *testing.T) {
	q := queue.New[int]()
	defer q.Close()

	// Push 50 items.
	for i := 0; i < 50; i++ {
		q.Push(i)
	}

	// And then pop 50 items, hoping to receive them in the correct order.
	for i := 0; i < 50; i++ {
		out := <-q.Pop()
		if out != i {
			t.Errorf("Queue output should be %d but is %d", i, out)
		}
	}

	// Check that there are no remaining items.
	select {
	case <-q.Pop():
		t.Error("Queue.Pop returned value")
	default:
	}
}

func TestQueuePointers(t *testing.T) {
	type Type struct {
		Value int
	}

	q := queue.New[*Type]()
	defer q.Close()

	// Push 50 items.
	for i := 0; i < 50; i++ {
		q.Push(&Type{Value: i})
	}

	// And then pop 50 items, hoping to receive them in the correct order.
	for i := 0; i < 50; i++ {
		out := <-q.Pop()
		if out.Value != i {
			t.Errorf("Queue output should be %d but is %d", i, out)
		}
	}

	// Check that there are no remaining items.
	select {
	case <-q.Pop():
		t.Error("Queue.Pop returned value")
	default:
	}
}

func TestQueueFlush(t *testing.T) {
	q := queue.New[int]()
	defer q.Close()
	timeout := time.NewTicker(time.Millisecond)
	defer timeout.Stop()

	// Flush.
	q.Flush()

	// Push some items.
	for i := 0; i < 3; i++ {
		q.Push(i)
	}

	// Check that queue is poppable.
	select {
	case _ = <-q.Pop():
	case <-timeout.C:
		t.Error("Queue.Pop did not return value")
	}

	// Flush.
	q.Flush()

	// Check that there are no remaining items.
	select {
	case <-q.Pop():
		t.Error("Queue.Pop returned value")
	case <-timeout.C:
	}
}

func TestQueueDiscard(t *testing.T) {
	q := queue.New[int]()
	defer q.Close()
	timeout := time.NewTicker(time.Millisecond)
	defer timeout.Stop()

	// Push an item.
	q.Push(1)

	// Start discarding.
	q.Discard(true)

	// Push some items.
	for i := 2; i < 5; i++ {
		q.Push(i)
	}

	// Stop discarding.
	q.Discard(false)

	// Push some more items.
	for i := 5; i < 7; i++ {
		q.Push(i)
	}

	// Check that queue is poppable and we only receive 1, 5, 6.
	count := 0
Poppable:
	for {
		select {
		case i := <-q.Pop():
			switch i {
			case 1, 5, 6:
				count++
				if count == 3 {
					break Poppable
				}
			default:
				t.Errorf("Queue.Pop did not return correct value: %d", i)
			}
		case <-timeout.C:
			t.Error("Queue.Pop did not return value")
		}
	}

	// Flush.
	q.Flush()

	// Check that there are no remaining items.
	select {
	case <-q.Pop():
		t.Error("Queue.Pop returned value")
	case <-timeout.C:
	}
}

func BenchmarkQueuePush(b *testing.B) {
	q := queue.New[int]()
	defer q.Close()
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
}

func BenchmarkQueuePop(b *testing.B) {
	q := queue.New[int]()
	defer q.Close()
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-q.Pop()
	}
}

func BenchmarkChannelPush(b *testing.B) {
	q := make(chan string, b.N)
	item := "item"

	for i := 0; i < b.N; i++ {
		q <- item
	}
}

func BenchmarkChannelPop(b *testing.B) {
	q := make(chan string, b.N)
	item := "item"

	for i := 0; i < b.N; i++ {
		q <- item
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-q
	}
}

func ExampleQueue() {
	q := queue.New[string]()
	defer q.Close()

	q.Push("item")
	out := <-q.Pop()
	fmt.Printf("received: %v\n", out)

	// Output:
	// received: item
}
