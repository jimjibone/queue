package queue_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jimjibone/queue"
)

func TestQueueSimple(t *testing.T) {
	q := queue.New()
	defer q.Close()

	item := "item"
	q.Push(item)
	out := <-q.Pop()
	if output, ok := out.(string); !ok {
		t.Errorf("Queue out type should be string but is %T", out)
	} else if output != item {
		t.Errorf("Queue output should be %q but is %q", item, output)
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
	q := queue.New()
	defer q.Close()

	// Push 50 items.
	for i := 0; i < 50; i++ {
		q.Push(i)
	}

	// And then pop 50 items, hoping to receive them in the correct order.
	for i := 0; i < 50; i++ {
		o := <-q.Pop()
		if outVal, ok := o.(int); !ok {
			t.Errorf("Queue output type should be int but is %T", o)
		} else if outVal != i {
			t.Errorf("Queue output should be %d but is %d", i, outVal)
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
	q := queue.New()
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

func BenchmarkQueuePush(b *testing.B) {
	q := queue.New()
	defer q.Close()
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
}

func BenchmarkQueuePop(b *testing.B) {
	q := queue.New()
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
	q := queue.New()
	defer q.Close()

	q.Push("item")
	out := <-q.Pop()
	fmt.Printf("received: %v\n", out)

	// Output:
	// received: item
}
