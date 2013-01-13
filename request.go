package requester

import (
	"log"
	"net/http"
)

var curId int

type Action struct {
	Id    int
	Req   *http.Request
	Retry int
	Body  string
}

func Request(req *http.Request) {
	curId++
	action := &Action{
		Id:  curId,
		Req: req,
	}

	actionsLogger.Printf("[%d] Enqueue request to %s", action.Id, req.URL.String())

	waitQueue.Add(1)
	go enqueueAction(action)
}

func GET(url string) *http.Request {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	return r
}

func WaitEmptyQueue() {
	waitQueue.Wait()
	shutdownData <- true
	<-waitData
	log.Println("Work finished!")
}
