package queue_test

import (
	"fmt"
	"testing"

	"github.com/jimjibone/queue"
)

func TestPubSubSimple(t *testing.T) {
	pub := queue.NewPub()
	defer pub.Close()

	sub1 := pub.NewSub()
	sub2 := pub.NewSub()
	defer sub1.Close()
	defer sub2.Close()

	item := "item"
	pub.Pub(item)

	out := <-sub1.Sub()
	if output, ok := out.(string); !ok {
		t.Errorf("Sub out type should be string but is %T", out)
	} else if output != item {
		t.Errorf("Sub output should be %q but is %q", item, output)
	}

	out = <-sub2.Sub()
	if output, ok := out.(string); !ok {
		t.Errorf("Sub out type should be string but is %T", out)
	} else if output != item {
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
	pub := queue.NewPub()
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
