package requester

import (
	"sync"
)

var (
	queue     = []*Request{}
	queueCh   = make(chan *Request, 500)
	curId     = 0
	waitQueue = &sync.WaitGroup{}
)

func addQueue(req *Request) {
	if req.Id == 0 {
		curId++
		req.Id = curId

		waitQueue.Add(1)
	}

	queueCh <- req
}

func popQueue() *Request {
	return <-queueCh
}
