package requester

import (
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func cachedResponse(req *Request) (*Response, bool) {
	f, err := os.Open(filepath.Join("cache", cacheName(req)))
	if err != nil {
		if !os.IsNotExist(err) {
			errLogger.Printf("[%d] Cache read failed [%s]: %s\n", req.Id, err)
		}
		return nil, false
	}
	defer f.Close()

	resp := &Response{}
	if err := gob.NewDecoder(f).Decode(resp); err != nil {
		errLogger.Printf("[%d] Cache decoding failed [%s]: %s\n", req.Id, err)
	}

	actionsLogger.Printf("[%d] Request read from cache: %s \n", req.Id,
		cacheName(req))

	return resp, true
}

func saveCache(req *Request, resp *Response) {
	f, err := os.Create(filepath.Join("cache", cacheName(req)))
	if err != nil {
		errLogger.Printf("[%d] Cache write failed [%s]: %s\n", req.Id, err)
		return
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(resp); err != nil {
		errLogger.Printf("[%d] Cache encoding failed [%s]: %s\n", req.Id, err)
	}
}

func deleteCache(req *Request) {
	if err := os.Remove(filepath.Join("cache", cacheName(req))); err != nil {
		if os.IsNotExist(err) {
			return
		}

		errLogger.Printf("[%d] Cache delete failed [%s]: %s\n", req.Id, err)
	}
}

func cacheName(req *Request) string {
	h := md5.New()
	io.WriteString(h, req.URL())
	return fmt.Sprintf("%x", h.Sum(nil))
}
