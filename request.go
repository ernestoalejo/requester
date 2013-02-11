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
}

func GET(url string) *Request {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// The URL was ill-formed. Exit directly, it's not a common error
		panic(err)
	}

	return &Request{Req: r}
}

func POST(url string, values url.Values) *Request {
	body := bytes.NewBuffer(nil)
	body.WriteString(values.Encode())

	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		// The URL was ill-formed. Exit directly, it's not a common error
		panic(err)
	}

	return &Request{Req: r}
}

func (r *Request) Send() {
	addQueue(r)
}

func (r *Request) URL() string {
	return r.Req.URL.String()
}
