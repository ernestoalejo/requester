package requester

import (
	"sync"
)

const (
	COUNTER_PROCESSED = "processed"
)

var (
	countersMutex = &sync.RWMutex{}
	counters      = map[string]*Counter{}
)

type Counter struct {
	mutex *sync.Mutex
	value int64
}

func (c *Counter) Increment() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value++
	return c.value
}

func (c *Counter) Decrement() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value--
	return c.value
}

func (c *Counter) Value() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.value
}

func GetCounter(name string) *Counter {
	countersMutex.RLock()
	c, ok := counters[name]
	countersMutex.RUnlock()

	if !ok {
		countersMutex.Lock()
		defer countersMutex.Unlock()

		if newc, ok := counters[name]; ok {
			return newc
		}
		c = &Counter{
			mutex: &sync.Mutex{},
		}
		counters[name] = c
	}

	return c
}
