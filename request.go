package requester

import (
	"bytes"
	"net/http"
	"net/url"
)

type Request struct {
	Id    int
	Req   *http.Request
	Retry int
	Data  interface{}
}

func GET(url string) *Request {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// The URL was ill-formed. Exit directly, it's not a common error
		panic(err)
	}

	return &Request{Req: r}
}

func POST(u string, values url.Values) *Request {
	body := bytes.NewBufferString(values.Encode())
	r, err := http.NewRequest("POST", u, body)
	if err != nil {
		// The URL was ill-formed. Exit directly, it's not a common error
		panic(err)
	}

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return &Request{Req: r}
}

func (r *Request) Send() {
	addQueue(r)
}

func (r *Request) URL() string {
	return r.Req.URL.String()
}

func (r *Request) Header(key, value string) {
	r.Req.Header.Set(key, value)
}
