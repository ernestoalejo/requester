package requester

import (
	"net/http"
)

type Request struct {
	Id    int
	Req   *http.Request
	Retry int
}

func GET(url string) *Request {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
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
