package requester

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"time"
)

var client = &http.Client{}

// Executed in parallel by the init module when starting the library
// It handles the request and iterates
// TODO: Wait for a task before handling it
func worker() {
	for {
		errWrapper(handleRequest(popQueue()))
	}
}

func handleRequest(req *Request) error {
	start := time.Now()

	resp, err := cachedResponse(req)
	if err != nil {
		return err
	}
	if resp == nil {
		resp = performRequest(req)
		if resp == nil {
			// Enqueued again for a later try
			return nil
		}

		if err := saveCache(req, resp); err != nil {
			return err
		}
	}

	processResponse(req, resp)

	if resp == nil {
		ns := time.Since(start).Nanoseconds()
		min := 1 * 60 * 1e9 / (config.MaxMinute * config.MaxSimultaneous)
		if ns < min {
			actionsLogger.Printf("[%d] Throttled %d ms", req.Id, (min-ns)/1e6)
			time.Sleep(time.Duration(min-ns) * time.Nanosecond)
		}
	}

	return nil
}

func performRequest(req *Request) *Response {
	actionsLogger.Printf("[%d] Make request...\n", req.Id)

	resp, err := client.Do(req.Req)
	if err != nil {
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		queueAgain(req, fmt.Errorf("req code not ok: %s", resp.Status))
		return nil
	}

	if config.LogNet {
		reqDump, err := httputil.DumpRequestOut(req.Req, config.LogBody)
		if err != nil {
			queueAgain(req, err)
			return nil
		}

		resDump, err := httputil.DumpResponse(resp, config.LogBody)
		if err != nil {
			queueAgain(req, err)
			return nil
		}

		s := "===================================================================="
		netLogger.Printf("$REQUEST [%d]$\n%s\n\n%s\n\n%s\n\n\n\n", req.Id,
			reqDump, s, resDump)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		queueAgain(req, err)
		return nil
	}

	actionsLogger.Printf("[%d] Request done!\n", req.Id)

	return &Response{Body: string(body)}
}

func queueAgain(req *Request, err error) {
	deleteCache(req)

	req.Retry++
	if req.Retry > config.MaxRetries {
		// TODO: Don't exit, ignore this entry
		log.Fatalf("[%d] Max retries reached [%s]\n", req.Id, req.URL())
	}

	secs := math.Pow(2, float64(req.Retry))*100 + float64(rand.Int()%1000)

	errLogger.Printf("[%d] Action failed [Retry %d] [%s]: %s\n", req.Id,
		req.Retry, req.URL(), err)
	actionsLogger.Printf("[%d] Retrying in %d milliseconds (Retry %d)...\n",
		req.Id, int(secs), req.Retry)

	go func() {
		time.Sleep(time.Duration(secs) * time.Millisecond)
		req.Send()
	}()
}

func processResponse(req *Request, resp *Response) {
	defer func() {
		if rec := recover(); rec != nil {
			queueAgain(req, fmt.Errorf("panic recovered error: %s", rec))
		}
	}()

	actionsLogger.Printf("[%d] Processing response... \n", req.Id)
	if err := config.Processor(req, resp); err != nil {
		queueAgain(req, err)
		return
	}
	actionsLogger.Printf("[%d] Processing done! \n", req.Id)

	processed := GetCounter(COUNTER_PROCESSED).Increment()
	log.Printf("[%d -> %d] Pending %d reqs in queue \n", req.Id,
		processed, len(queue))

	// TODO: Save data
	//saveDataRequest <- true
	waitQueue.Done()
}
