package requester

import (
	"sync"
)

var (
	queueCh   = make(chan *Request, 100)
	curId     = 0
	waitQueue = &sync.WaitGroup{}
)

func addQueue(req *Request) {
	if req.Id == 0 {
		curId++
		req.Id = curId
		waitQueue.Add(1)
	}
	GetCounter(COUNTER_PENDING).Increment()
	queueCh <- req
}

func popQueue() *Request {
	r := <-queueCh
	return r
}
