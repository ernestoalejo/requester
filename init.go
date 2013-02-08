package requester

import (
	"os"
)

func InitLibrary(c *Config) error {
	if c.MaxMinute < c.MaxSimultaneous {
		return Errorf("config not safe: max/min should be >= than simultaneous")
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
	for i := int64(0); i < c.MaxSimultaneous; i++ {
		go worker()
	}

	return nil
}

func CloseLibrary() error {
	return closeDB()
}
