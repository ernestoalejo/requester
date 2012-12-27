package requester

import (
	"log"
)

var manuals int

func ManualRevision(action *Action) {
	log.Printf("[%d] Manual revision required!\n", action.Id)
	manualLogger.Printf("[%d] Manual revision required!\n", action.Id)

	manuals++
	if manuals >= config.MaxReviews {
		log.Fatal("Too much manual reviews")
	}
}
