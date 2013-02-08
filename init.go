package requester

import (
	"log"
	"os"
)

func InitLibrary(c *Config) {
	if c.MaxMinute < c.MaxSimultaneous {
		log.Fatalf("config not safe: max/min should be >= than simultaneous")
	}

	if err := os.MkdirAll("cache", 0766); err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll("loggers", 0766); err != nil {
		log.Fatal(err)
	}

	if err := initLoggers(); err != nil {
		log.Fatal(err)
	}

	if err := initDB(); err != nil {
		log.Fatal(err)
	}

	config = c

	for i := int64(0); i < c.MaxSimultaneous; i++ {
		go worker()
	}
}

func CloseLibrary() {
	if err := closeDB(); err != nil {
		log.Fatal(err)
	}
}
