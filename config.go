package requester

var (
	config *Config
)

type ProcessorFunc func(r *Request, w *Response) error

type Config struct {
	// Max requests per minute the app can make to the server
	MaxMinute int64

	// Max simultaneos request the app can have opened at the same time
	// to the server
	MaxSimultaneous int64

	// The processor function
	Processor ProcessorFunc

	// Enables the logging of network requests and responses
	LogNet bool

	// Enables the logging of the body of the requests & responses
	LogBody bool

	// Max failed retries to obtain a page before ignoring it
	MaxRetries int

	// Max DB operations the program will save in memory before commiting
	// all of them
	BufferedOperations int
}
