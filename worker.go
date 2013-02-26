package requester

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

var (
	client = &http.Client{}
)

// Executed in parallel by the init module when starting the library
// It handles the request and iterates
func worker() {
	for {
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

	var reqDump []byte
	if config.LogNet {
		var err error
		reqDump, err = httputil.DumpRequestOut(req.Req, config.LogBody)
		if err != nil {
			return nil, Error(err)
		}
	}

	start := time.Now()

	resp, err := client.Do(req.Req)
	if err != nil {
		return nil, Error(err)
	}
	defer resp.Body.Close()

	if config.LogNet {
		resDump, err := httputil.DumpResponse(resp, config.LogBody)
		if err != nil {
			return nil, Error(err)
		}

		s := "===================================================================="
		netLogger.Printf("$REQUEST [%d]$\n%s\n\n%s\n\n%s\n\n\n\n", req.Id,
			reqDump, s, resDump)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, Errorf("req code not ok: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, Error(err)
	}

	ms := time.Since(start).Nanoseconds() / 1e6
	var reqtime string
	if ms > 1000 {
		reqtime = fmt.Sprintf("%.3f s", float64(ms)/1000.)
	} else {
		reqtime = fmt.Sprintf("%d ms", ms)
	}

	actionsLogger.Printf("[%d] Request done in %s!\n", req.Id, reqtime)

	return &Response{Body: convertUTF8(resp, string(body))}, nil
}

func queueAgain(req *Request, err error) error {
	if err := deleteCache(req); err != nil {
		return err
	}

	req.Retry++
	if req.Retry > config.MaxRetries {
		// TODO: Output the list of failed requests at the end (a new logger?)
		errLogger.Printf("[%d] Max retries reached [%s]\n", req.Id, req.URL())
		waitQueue.Done()
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
	pending := GetCounter(COUNTER_PENDING).Decrement()
	log.Printf("[%d -> %d] Pending %d reqs in queue \n", req.Id,
		processed, pending)

	waitQueue.Done()
	return nil
}

// Try to detect the encoding of the page. First trying to read a UTF-8
// encoding from the Content-Type header. Then it tries to read the <meta> tag
// in the first 1000 chars of the response. If all fails, it's supposed to
// use ISO-8859-1, and it's converted back to UTF-8.
func convertUTF8(resp *http.Response, s string) string {
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "charset=utf-8") {
		return s
	}

	body := strings.ToLower(s[:min(len(s), 10000)])
	if strings.Contains(body, "charset=utf-8") {
		return s
	}

	return UTF8(s)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
