package requester

import (
	"log"
)

func ManualRevision(action *Action) {
	log.Printf("[%d] Manual revision required!\n", action.Id)
	manualLogger.Printf("[%d] Manual revision required!\n", action.Id)
}
