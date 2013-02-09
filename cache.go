package requester

import (
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func cachedResponse(req *Request) (*Response, error) {
	f, err := os.Open(filepath.Join("cache", cacheName(req)))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, Error(err)
		}
		return nil, nil
	}
	defer f.Close()

	resp := &Response{}
	if err := gob.NewDecoder(f).Decode(resp); err != nil {
		return nil, Error(err)
	}

	actionsLogger.Printf("[%d] Request read from cache: %s \n", req.Id,
		cacheName(req))

	return resp, nil
}

func saveCache(req *Request, resp *Response) error {
	f, err := os.Create(filepath.Join("cache", cacheName(req)))
	if err != nil {
		return Error(err)
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(resp); err != nil {
		return Error(err)
	}

	return nil
}

func deleteCache(req *Request) error {
	if err := os.Remove(filepath.Join("cache", cacheName(req))); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return Error(err)
	}
	return nil
}

func cacheName(req *Request) string {
	h := md5.New()
	io.WriteString(h, req.URL())
	return fmt.Sprintf("%x", h.Sum(nil))
}
