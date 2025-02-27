package ring_buffer

import (
	"fmt"
	"sync"
	"testing"
)

type queueInterface[T comparable] interface {
	Dequeue() T
	Enqueue(item T)
}

func TestConcurrentQueue(t *testing.T) {
	queue := New[int]()
	totalItems := 100
	var wg sync.WaitGroup
	wg.Add(2)
	go produce(queue, &wg, totalItems)
	go consume(queue, &wg, totalItems)
	wg.Wait()
}

func TestAtomicQueue(t *testing.T) {
	queue := NewAtomicQueue[int]()
	totalItems := 100
	var wg sync.WaitGroup
	wg.Add(2)
	go produce(queue, &wg, totalItems)
	go consume(queue, &wg, totalItems)
	wg.Wait()
}

func produce(q queueInterface[int], wg *sync.WaitGroup, totalItemsToQueue int) {
	i := 1
	for i <= totalItemsToQueue {
		q.Enqueue(i)
		fmt.Println("put: ", i)
		i++
	}
	wg.Done()
}

var result int

func consume(q queueInterface[int], wg *sync.WaitGroup, totalItemsFromQ int) {
	var item int
	i := 1

	for i <= totalItemsFromQ {
		item = q.Dequeue()
		fmt.Println("Get: ", item)
		i++
	}
	result = item
	wg.Done()
}
