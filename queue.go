package requester

import (
	"sync"
)

var (
	queue      = []*Request{}
	queueMutex = &sync.Mutex{}
	curId      = 0
	waitQueue  = &sync.WaitGroup{}
	started    = false
)

func addQueue(req *Request) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if started {
		panic("request added after the work has started")
	}

	if req.Id == 0 {
		curId++
		req.Id = curId
		waitQueue.Add(1)
	}
	queue = append(queue, req)
}

func initQueue() {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if started {
		panic("restarting twice the work")
	}
	started = true

	m := config.MaxSimultaneous
	if m > int64(len(queue)) {
		m = int64(len(queue))
	}

	for i := int64(0); i < m; i++ {
		go worker()
	}
}

func popQueue() *Request {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if len(queue) == 0 {
		return nil
	}

	var r *Request
	r, queue = queue[len(queue)-1], queue[:len(queue)-1]
	return r
}
