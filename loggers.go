package requester

import (
	"log"
	"os"
)

var errLogger, actionsLogger, netLogger *log.Logger

func initLoggers() error {
	f, err := os.Create("errors.log")
	if err != nil {
		return err
	}
	errLogger = log.New(f, "", log.Ldate|log.Ltime|log.Lshortfile)

	f, err = os.Create("actions.log")
	if err != nil {
		return err
	}
	actionsLogger = log.New(f, "", log.Ldate|log.Ltime)

	f, err = os.Create("net.log")
	if err != nil {
		return err
	}
	netLogger = log.New(f, "", log.Ldate|log.Ltime)

	return nil
}
