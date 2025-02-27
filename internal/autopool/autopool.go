// Inspired by John Silver
// https://github.com/johnsiilver/golib/blob/master/development/autopool/blend/autopool.go

package autopool

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

type stat struct {
	total, misses uint64
}

func (s *stat) miss() {
	atomic.AddUint64(&s.misses, 1)
}

func (s *stat) incr() {
	atomic.AddUint64(&s.total, 1)
}

// Pool provides a sync.Pool where GC'd objects are put back into the Pool automatically.
type Pool struct {
	stats []*stat

	typeToInt map[reflect.Type]int
	pools     []*sync.Pool
	mu        sync.Mutex
}

// New is the constructor for Pool.
func New() *Pool {
	return &Pool{pools: []*sync.Pool{}, typeToInt: map[reflect.Type]int{}}
}

// Add adds support for type returned by f(). This must be a pointer to a struct.
// Calling this after the first call to Get() or Put() is a race condition.
func (p *Pool) Add(f func() interface{}) int {
	i := f()
	t := reflect.TypeOf(i)

	switch t.Kind() {
	case reflect.Ptr:
	default:
		panic(fmt.Sprintf("cannot create a Pool on %v value", t))
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if v, ok := p.typeToInt[t]; ok {
		return v
	}

	p.stats = append(p.stats, new(stat))
	pos := len(p.stats) - 1

	p.typeToInt[t] = len(p.pools)
	p.pools = append(
		p.pools,
		&sync.Pool{
			New: func() interface{} {
				p.stats[pos].miss()
				return f()
			},
		},
	)
	return p.typeToInt[t]
}

// IntLookup will return the id used in Get() to fetch a value of this type. Returns -1 if the
// Pool does not suppor the type.
func (p *Pool) IntLookup(t reflect.Type) int {
	if v, ok := p.typeToInt[t]; ok {
		return v
	}
	return -1
}

// Get retrieves a value from a pool that has the type reflect.Type. This will panic if
// that type is an unsupported type or was not a sub type of the value that was provided by the
// function passed to New().
func (p *Pool) Get(i int) interface{} {
	p.stats[i].incr()

	pool := p.pools[i]

	x := pool.Get()
	runtime.SetFinalizer(x, func(o interface{}) { p.put(i, o) })

	// Note: The finalizer above uses the single goroutine that runs all finalizers.
	// It does not spin off a goroutine, as benchmarking finds that this actually adds a slight cost.
	// I tried a version that used a 1000 deep buffered channel with a 4 goroutine pool, but that
	// actually destroyed performance (almost eliminating any benefit). So the trivial cost of this
	// executing in a single tight loop seems to be hard to beat without spending a lot more resources.
	// It does mean that the code is susceptible to hijacking by other finalizers.

	return x
}

// Stats returns the number of misses (when we didn't have anything in a pool) and total attempts for
// each type we are storing. So if Add(<type>) returned 10, Stats()[10] will return a [2]uint64,
// with position 0 storing the misses and position 1 the total.
func (p *Pool) Stats() [][2]uint64 {
	out := make([][2]uint64, 0, len(p.stats))
	for _, stat := range p.stats {
		out = append(out, [2]uint64{atomic.LoadUint64(&stat.misses), atomic.LoadUint64(&stat.total)})
	}
	return out
}

// put adds x to the pool.
func (p *Pool) put(i int, x interface{}) {
	p.pools[i].Put(x)
}
