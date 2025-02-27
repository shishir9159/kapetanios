package ring_buffer

import (
	"sync"
)

type ConcurrentQueue[T comparable] struct {
	items []T
	lock  sync.Mutex
	cond  *sync.Cond
}

func New[T comparable]() *ConcurrentQueue[T] {
	q := &ConcurrentQueue[T]{}
	q.cond = sync.NewCond(&q.lock)
	return q
}

func (q *ConcurrentQueue[T]) Enqueue(item T) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.items = append(q.items, item)
	q.cond.Signal()
}

func (q *ConcurrentQueue[T]) Dequeue() T {
	q.lock.Lock()
	defer q.lock.Unlock()
	// if Get is called before Put,
	// then cond waits until the Put signals.
	for len(q.items) == 0 {
		q.cond.Wait()
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *ConcurrentQueue[T]) IsEmpty() bool {
	return len(q.items) == 0
}
