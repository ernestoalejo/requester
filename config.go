package requester

var (
	config *Config
)

type ProcessorFunc func(response string) error

type Config struct {
	// Max request / minute the app can make to the server
	MaxMinute int

	// Max simultaneos request the app can have opened
	// to the server
	MaxSimultaneous int

	// True to avoid having a mutex before entering the processor
	ThreadSafe bool

	// The processor function
	Processor ProcessorFunc

	// Whether to log the request and response body from the server in the HTTP log
	LogBody bool

	// Max failed retries to obtain a page before exiting the program
	MaxRetries int
}

func ApplyConfig(c *Config) {
	config = c
}
