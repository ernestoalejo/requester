package requester

import (
	"os"
)

func InitLibrary() error {
	if err := initLoggers(); err != nil {
		return err
	}

	if err := os.MkdirAll("cache", 0766); err != nil {
		return err
	}

	go handler()

	return nil
}
