package requester

import (
	"os"
)

func InitLibrary(c *Config) error {
	if c.MaxMinute < c.MaxSimultaneous {
		return Errorf("config not safe: max/min should be >= than simultaneous")
	}
	if !c.LogNet && c.LogBody {
		return Errorf("cannot log the body of requests if net logger is not enabled")
	}

	if err := os.MkdirAll("cache", 0766); err != nil {
		return Error(err)
	}
	if err := os.MkdirAll("loggers", 0766); err != nil {
		return Error(err)
	}

	if err := initLoggers(); err != nil {
		return err
	}
	if err := initDB(); err != nil {
		return err
	}

	config = c

	workerCh = make(chan bool, c.MaxSimultaneous)
	for i := int64(0); i < c.MaxSimultaneous; i++ {
		go worker()
	}

	return nil
}

func CloseLibrary() error {
	waitQueue.Wait()
	return closeDB()
}
