# queue

A basic event queue (and publisher/subscriber) in go, now supporting Go v1.18+ Generics in `github.com/jimjibone/queue/v2`.

See the [`v1` readme](../v1/README.md) for the older interface based implementation.

## Installation

```sh
go get github.com/jimjibone/queue/v2
```

## Queue Usage

Queue is a channel-based FIFO queue. Similar to a Go channel, items can be
pushed to the back of the Queue and then popped off the front by listening on
the Pop channel. This structure differs from channels in that its buffer is
effectively endless.

```go
q := queue.New[string]()
defer q.Close()

q.Push("item")
out := <-q.Pop()
fmt.Printf("received: %v\n", out)
```

## Pub-Sub Usage

Pub is a channel-based broadcast FIFO queue. Built on top of Queue, the Pub
sends published messages to all active Subs created via the NewSub method.

```go
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
```
