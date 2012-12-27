package requester

var (
	config *Config
)

type ProcessorFunc func(response string) error

type Config struct {
	// Max request / minute the app can make to the server
	MaxMinute int64

	// Max simultaneos request the app can have opened
	// to the server
	MaxSimultaneous int64

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
	if queueCount != 0 {
		panic("cannot change config while it's running")
	}
	config = c

	if config.MaxMinute < config.MaxSimultaneous {
		panic("config not safe: max/min should be equal or greater than simultaneous")
	}

	slot = make(chan bool, config.MaxSimultaneous)
	for i := int64(0); i < config.MaxSimultaneous; i++ {
		slot <- true
	}
}
