package requester

var (
	config *Config
)

type ProcessorFunc func(action *Action) error

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

	// Max number of process that can be marked as manual review
	// before exiting the app
	MaxReviews int

	// Global data to save periodically to disk
	AppData interface{}

	loaded bool
}
