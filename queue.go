package requester

import (
	"sync"
)

var (
	queue      = []*Request{}
	queueMutex = &sync.Mutex{}
	waitQueue  = sync.WaitGroup{}
	curId      = 0
	queueCh    = make(chan bool)
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
	case queueCh <- true:
	default:
	}
}

func popQueue() *Request {
	r := getQueue()
	for r == nil {
		<-queueCh
		r = getQueue()
	}
	return r
}

func getQueue() *Request {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if len(queue) == 0 {
		return nil
	}

	var req *Request
	req, queue = queue[0], queue[1:]
	return req
}
