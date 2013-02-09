package requester

import (
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"time"
)

var (
	client   = &http.Client{}
	workerCh chan bool
)

// Executed in parallel by the init module when starting the library
// It handles the request and iterates
func worker() {
	for {
		<-workerCh
		req := popQueue()
		errWrapper(req, handleRequest(req))
	}
}

func handleRequest(req *Request) error {
	start := time.Now()

	resp, err := cachedResponse(req)
	if err != nil {
		return err
	}
	if resp == nil {
		resp, err = performRequest(req)
		if err != nil {
			if err := queueAgain(req, err); err != nil {
				return err
			}
			return nil
		}
		if err := saveCache(req, resp); err != nil {
			return err
		}
	}

	if err := processResponse(req, resp); err != nil {
		if err := queueAgain(req, err); err != nil {
			return err
		}
		return nil
	}

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

func performRequest(req *Request) (*Response, error) {
	actionsLogger.Printf("[%d] Make request...\n", req.Id)

	resp, err := client.Do(req.Req)
	if err != nil {
		return nil, Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, Errorf("req code not ok: %s", resp.Status)
	}

	if config.LogNet {
		reqDump, err := httputil.DumpRequestOut(req.Req, config.LogBody)
		if err != nil {
			return nil, Error(err)
		}

		resDump, err := httputil.DumpResponse(resp, config.LogBody)
		if err != nil {
			return nil, Error(err)
		}

		s := "===================================================================="
		netLogger.Printf("$REQUEST [%d]$\n%s\n\n%s\n\n%s\n\n\n\n", req.Id,
			reqDump, s, resDump)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, Error(err)
	}

	actionsLogger.Printf("[%d] Request done!\n", req.Id)

	return &Response{Body: string(body)}, nil
}

func queueAgain(req *Request, err error) error {
	if err := deleteCache(req); err != nil {
		return err
	}

	req.Retry++
	if req.Retry > config.MaxRetries {
		// TODO: Output the list of failed requests at the end (a new logger?)
		errLogger.Printf("[%d] Max retries reached [%s]\n", req.Id, req.URL())
		return nil
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

	return nil
}

func processResponse(req *Request, resp *Response) (reterr error) {
	defer func() {
		if rec := recover(); rec != nil {
			reterr = Errorf("panic recovered error: %s", rec)
		}
	}()

	actionsLogger.Printf("[%d] Processing response... \n", req.Id)
	if err := config.Processor(req, resp); err != nil {
		return err
	}
	actionsLogger.Printf("[%d] Processing done! \n", req.Id)

	processed := GetCounter(COUNTER_PROCESSED).Increment()
	log.Printf("[%d -> %d] Pending %d reqs in queue \n", req.Id,
		processed, len(queue))

	waitQueue.Done()
	return nil
}
