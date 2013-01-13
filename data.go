package requester

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var (
	saveDataRequest = make(chan bool, 100)

	shutdownData = make(chan bool)
	waitData     = make(chan bool)
)

func data() {
	r := 0
	ticker := time.NewTicker(3 * time.Second)
	exit := false
	for !exit {
		select {
		case <-saveDataRequest:
			r++
			continue

		case <-ticker.C:
			if r == 0 {
				continue
			}

		case <-shutdownData:
			exit = true
		}

		log.Printf("Saving global data [%d requests]...\n", r)
		r = 0
		saveData()
	}

	ticker.Stop()
	waitData <- true
}

func LoadData() {
	if config == nil {
		return
	}
	if config.loaded {
		return
	}
	config.loaded = true

	f, err := os.Open("db")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}

		log.Fatal(err)
	}
	if err := json.NewDecoder(f).Decode(config.AppData); err != nil {
		log.Fatal(err)
	}
	f.Close()
}

func saveData() {
	f, err := os.Create("db")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(config.AppData); err != nil {
		log.Fatal(err)
	}
	f.Close()
}
