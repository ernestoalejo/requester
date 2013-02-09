package requester

import (
	"sync"
)

var (
	queue      = []*Request{}
	queueMutex = &sync.Mutex{}
	waitQueue  = sync.WaitGroup{}
	curId      = 0
)

func addQueue(req *Request) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if req.Id == 0 {
		curId++
		req.Id = curId

		waitQueue.Add(1)
	}

	queue = append(queue, req)

	// Non-blocking notification for a worker to start processing
	// the request
	select {
	case workerCh <- true:
	default:
	}
}

func popQueue() *Request {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	var req *Request
	req, queue = queue[0], queue[1:]
	return req
}
