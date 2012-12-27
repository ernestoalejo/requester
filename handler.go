package requester

import (
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"time"
)

var (
	queue  = make(chan *Action, 50)
	wait   = make(chan bool)
	client = &http.Client{}

	queueMutex = &sync.Mutex{}
	queueCount = 0

	processMutex = &sync.Mutex{}
	processed    = 0
)

func InitLibrary() error {
	if err := initLoggers(); err != nil {
		return err
	}

	if err := os.MkdirAll("cache", 0766); err != nil {
		return err
	}

	go handler()

	return nil
}

func handler() {
	for {
		action := <-queue

		go func() {
			if !cache(action) {
				perform(action)
				saveCache(action)
			}
			process(action)
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

	reqDump, err := httputil.DumpRequestOut(action.Req, config.LogBody)
	if err != nil {
		log.Fatal(err)
	}

	resDump, err := httputil.DumpResponse(resp, config.LogBody)
	if err != nil {
		log.Fatal(err)
	}

	s := "===================================================================="

	netLogger.Printf("$REQUEST$\n%s\n\n%s\n\n%s\n\n\n\n", reqDump, s, resDump)

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
	if err := config.Processor(action.Body); err != nil {
		queueAgain(action, err)
	}
	actionsLogger.Printf("[%d] Processing done! \n", action.Id)

	processed++
	log.Printf("[%d] Pending %d requests in queue \n", processed, queueCount)

	queueMutex.Lock()
	queueCount--
	if queueCount == 0 {
		wait <- true
	}
	queueMutex.Unlock()
}

func queueAgain(action *Action, err error) {
	action.Retry++
	if action.Retry > config.MaxRetries {
		log.Fatalf("[%d] Max retries reached [%s]\n", action.Id,
			action.Req.URL.String())
	}

	errLogger.Printf("[%d] Action failed [Retry %d] [%s]: %s\n", action.Id,
		action.Retry, action.Req.URL.String(), err)

	queueCount++
	go func() {
		secs := math.Pow(2, float64(action.Retry))*100 + float64(rand.Int()%1000)

		actionsLogger.Printf("[%d] Retrying in %d milliseconds (Retry %d)...\n",
			action.Id, int(secs), action.Retry)

		time.Sleep(time.Duration(secs) * time.Millisecond)
		queue <- action
	}()
}
