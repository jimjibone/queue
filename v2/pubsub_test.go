package queue_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jimjibone/queue/v2"
)

func TestPubSubSimple(t *testing.T) {
	pub := queue.NewPub[string]()
	// defer pub.Close() -- test note: closing before sub2 is closed to cover extra cases.

	sub1 := pub.NewSub()
	sub2 := pub.NewSub()
	defer sub1.Close()
	defer pub.Close()
	defer sub2.Close()

	item := "item"
	pub.Pub(item)

	output := <-sub1.Sub()
	if output != item {
		t.Errorf("Sub output should be %q but is %q", item, output)
	}

	output = <-sub2.Sub()
	if output != item {
		t.Errorf("Sub output should be %q but is %q", item, output)
	}

	select {
	case <-sub1.Sub():
		t.Error("sub1.Sub returned value")
	case <-sub2.Sub():
		t.Error("sub2.Sub returned value")
	default:
	}

	// Test closing while there are still items in the Subs.
	pub.Pub("item2")
}

func ExamplePub() {
	pub := queue.NewPub[string]()
	defer pub.Close()

	sub1 := pub.NewSub()
	sub2 := pub.NewSub()
	defer sub1.Close()
	defer sub2.Close()

	pub.Pub("item")
	out1 := <-sub1.Sub()
	fmt.Printf("sub1 received: %v\n", out1)
	out2 := <-sub2.Sub()
	fmt.Printf("sub2 received: %v\n", out2)

	// Output:
	// sub1 received: item
	// sub2 received: item
}

// The shape some servers use: a run goroutine replays current state to a newly
// registered Sub with Send, while that Sub's owner closes it. Send must not
// strand the run goroutine, because that goroutine also serves every other
// subscriber and its own shutdown.
func TestSendToClosedSubDoesNotBlock(t *testing.T) {
	pub := queue.NewPub[int]()
	defer pub.Close()

	sub := pub.NewSub()
	sub.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			pub.Send(sub, i)
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Send blocked on a closed Sub")
	}
}

// As above, but with the close racing the replay.
func TestSendDuringSubCloseDoesNotBlock(t *testing.T) {
	for i := 0; i < 50; i++ {
		pub := queue.NewPub[int]()
		sub := pub.NewSub()

		replayed := make(chan struct{})
		go func() {
			defer close(replayed)
			// A snapshot big enough that the close lands mid-replay.
			for j := 0; j < 200; j++ {
				pub.Send(sub, j)
			}
		}()

		sub.Close()

		select {
		case <-replayed:
		case <-time.After(2 * time.Second):
			t.Fatal("Send blocked while the Sub was closing")
		}
		pub.Close()
	}
}

// Pub must stay unaffected: a closed subscriber is skipped, live ones still
// receive.
func TestPubSkipsClosedSub(t *testing.T) {
	pub := queue.NewPub[string]()
	defer pub.Close()

	closed := pub.NewSub()
	live := pub.NewSub()
	defer live.Close()
	closed.Close()

	pub.Pub("item")

	select {
	case got := <-live.Sub():
		if got != "item" {
			t.Errorf("live sub got %q, want %q", got, "item")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Pub did not reach the live subscriber")
	}
}
