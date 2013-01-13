package requester

import (
	"fmt"
	"os"
)

func InitLibrary(c *Config) error {
	if config != nil {
		return fmt.Errorf("cannot set the config twice")
	}
	if c.MaxMinute < c.MaxSimultaneous {
		return fmt.Errorf("config not safe: max/min should be >= than simultaneous")
	}

	if err := initLoggers(); err != nil {
		return err
	}

	if err := os.MkdirAll("cache", 0766); err != nil {
		return err
	}

	config = c
	LoadData()

	go data()
	for i := int64(0); i < c.MaxSimultaneous; i++ {
		go worker()
	}

	return nil
}
