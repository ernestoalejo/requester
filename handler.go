package requester

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

var (
	queue  = make(chan *Action, 50)
	slot   chan bool
	client = &http.Client{}

	processMutex = &sync.Mutex{}

	wait     = make(chan bool)
	waitData = make(chan bool)

	saveRequest = make(chan bool, 100)

	waitProcessed = sync.WaitGroup{}
)

func handler() {
	for {
		<-slot

		action := <-queue

		go func() {
			start := time.Now()

			cached := true
			if !cache(action) {
				perform(action)
				saveCache(action)
				cached = false
			}
			process(action)

			if !cached {
				ns := time.Since(start).Nanoseconds()
				min := 1 * 60 * 1e9 / config.MaxMinute
				if ns < min {
					actionsLogger.Printf("[%d] Throttled %d ms", action.Id,
						(min-ns)/1000000)
					time.Sleep(time.Duration(min-ns) * time.Nanosecond)
				}
			}
			slot <- true
		}()
	}
}

func perform(action *Action) {
	actionsLogger.Printf("[%d] Make request...\n", action.Id)

	resp, err := client.Do(action.Req)
	if err != nil {
		queueAgain(action, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		queueAgain(action, fmt.Errorf("action code not ok: %s", resp.Status))
		return
	}

	reqDump, err := httputil.DumpRequestOut(action.Req, config.LogBody)
	if err != nil {
		log.Fatal(err)
	}

	resDump, err := httputil.DumpResponse(resp, config.LogBody)
	if err != nil {
		log.Fatal(err)
	}

	s := "===================================================================="

	netLogger.Printf("$REQUEST [%d]$\n%s\n\n%s\n\n%s\n\n\n\n", action.Id,
		reqDump, s, resDump)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	action.Body = string(body)

	actionsLogger.Printf("[%d] Request done!\n", action.Id)
}

func process(action *Action) {
	if !config.ThreadSafe {
		processMutex.Lock()
		defer processMutex.Unlock()
	}

	actionsLogger.Printf("[%d] Processing response... \n", action.Id)
	if err := config.Processor(action); err != nil {
		queueAgain(action, err)
	}
	actionsLogger.Printf("[%d] Processing done! \n", action.Id)

	processed := GetCounter(COUNTER_PROCESSED).Increment()
	queueCount := GetCounter(COUNTER_REQUESTS).Value()
	log.Printf("[%d -> %d] Pending %d requests in queue \n", action.Id,
		processed, queueCount)

	queueCount = GetCounter(COUNTER_REQUESTS).Decrement()
	if queueCount == 0 {
		waitData <- true
	}
	saveRequest <- true
}

func queueAgain(action *Action, err error) {
	action.Retry++
	if action.Retry > config.MaxRetries {
		log.Fatalf("[%d] Max retries reached [%s]\n", action.Id,
			action.Req.URL.String())
	}

	errLogger.Printf("[%d] Action failed [Retry %d] [%s]: %s\n", action.Id,
		action.Retry, action.Req.URL.String(), err)

	GetCounter(COUNTER_REQUESTS).Increment()
	go func() {
		secs := math.Pow(2, float64(action.Retry))*100 + float64(rand.Int()%1000)

		actionsLogger.Printf("[%d] Retrying in %d milliseconds (Retry %d)...\n",
			action.Id, int(secs), action.Retry)

		time.Sleep(time.Duration(secs) * time.Millisecond)
		queue <- action
	}()
}
