package requester

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func cache(action *Action) bool {
	name := buildCacheName(action.Req)
	f, err := os.Open(filepath.Join("cache", name))
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Fatal(err)
	}
	defer f.Close()

	read, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	action.Body = string(read)

	actionsLogger.Printf("[%d] Request read from cache: %s \n", action.Id, name)

	return true
}

func saveCache(action *Action) {
	name := buildCacheName(action.Req)
	f, err := os.Create(filepath.Join("cache", name))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Fprint(f, action.Body)
}

func deleteCache(action *Action) {
	name := buildCacheName(action.Req)
	if err := os.Remove(filepath.Join("cache", name)); err != nil {
		if os.IsNotExist(err) {
			return
		}

		log.Fatal(err)
	}
}

func buildCacheName(req *http.Request) string {
	h := md5.New()
	io.WriteString(h, req.URL.String())
	return fmt.Sprintf("%x", h.Sum(nil))
}
