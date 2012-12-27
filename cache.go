package requester

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func cache(action *Action) bool {
	name := buildCacheName(action.Req.URL.String())
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

	actionsLogger.Printf("[%d] Request read from cache \n", action.Id)

	return true
}

func saveCache(action *Action) {
	name := buildCacheName(action.Req.URL.String())
	f, err := os.Create(filepath.Join("cache", name))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Fprint(f, action.Body)
}

func buildCacheName(source string) string {
	h := md5.New()
	io.WriteString(h, source)
	return fmt.Sprintf("%x", h.Sum(nil))
}
