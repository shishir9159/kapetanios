package ring_buffer

import (
	"sync/atomic"
)

// by using channel, we can make sure the queue is thread safe,
// as only one goroutine can send or receive from a channel
// at a time. Atomic operation ensures no data race condition happen
// as counter variable would only be updated as sequential manner

type AtomicQueue[T comparable] struct {
	items   chan T
	counter uint64
}

func NewAtomicQueue[T comparable]() *AtomicQueue[T] {
	return &AtomicQueue[T]{
		items:   make(chan T, 1),
		counter: 0,
	}
}

func (q *AtomicQueue[T]) Enqueue(item T) {
	// auto increment
	atomic.AddUint64(&q.counter, 1)
	q.items <- item
}

func (q *AtomicQueue[T]) Dequeue() T {
	item := <-q.items
	// counter variable decremented atomically
	atomic.AddUint64(&q.counter, ^uint64(0))
	return item
}

func (q *AtomicQueue[T]) IsEmpty() bool {
	return q.counter == 0
}
