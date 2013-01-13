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
	queue        = make(chan *Action)
	client       = &http.Client{}
	processMutex = &sync.Mutex{}
	waitQueue    = sync.WaitGroup{}
)

func worker() {
	for {
		action := <-queue
		handleAction(action)
	}
}

func handleAction(action *Action) {
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
			actionsLogger.Printf("[%d] Throttled %d ms", action.Id, (min-ns)/1e6)
			time.Sleep(time.Duration(min-ns) * time.Nanosecond)
		}
	}
}

func perform(action *Action) {
	actionsLogger.Printf("[%d] Make request...\n", action.Id)

	resp, err := client.Do(action.Req)
	if err != nil {
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		queueAgain(action, fmt.Errorf("action code not ok: %s", resp.Status))
		return
	}

	reqDump, err := httputil.DumpRequestOut(action.Req, config.LogBody)
	if err != nil {
		queueAgain(action, err)
		return
	}

	resDump, err := httputil.DumpResponse(resp, config.LogBody)
	if err != nil {
		queueAgain(action, err)
		return
	}

	s := "===================================================================="

	netLogger.Printf("$REQUEST [%d]$\n%s\n\n%s\n\n%s\n\n\n\n", action.Id,
		reqDump, s, resDump)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		queueAgain(action, err)
		return
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
		return
	}
	actionsLogger.Printf("[%d] Processing done! \n", action.Id)

	processed := GetCounter(COUNTER_PROCESSED).Increment()
	queueCount := GetCounter(COUNTER_REQUESTS).Decrement()
	log.Printf("[%d -> %d] Pending %d requests in queue \n", action.Id,
		processed, queueCount)

	saveDataRequest <- true
	waitQueue.Done()
}

func queueAgain(action *Action, err error) {
	deleteCache(action)

	action.Retry++
	if action.Retry > config.MaxRetries {
		log.Fatalf("[%d] Max retries reached [%s]\n", action.Id,
			action.Req.URL.String())
	}

	secs := math.Pow(2, float64(action.Retry))*100 + float64(rand.Int()%1000)

	errLogger.Printf("[%d] Action failed [Retry %d] [%s]: %s\n", action.Id,
		action.Retry, action.Req.URL.String(), err)
	actionsLogger.Printf("[%d] Retrying in %d milliseconds (Retry %d)...\n",
		action.Id, int(secs), action.Retry)

	go func() {
		time.Sleep(time.Duration(secs) * time.Millisecond)
		enqueueAction(action)
	}()
}

func enqueueAction(action *Action) {
	GetCounter(COUNTER_REQUESTS).Increment()
	queue <- action
}
