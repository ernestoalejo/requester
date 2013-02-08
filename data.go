package requester

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db           *sql.DB
	tx           *sql.Tx
	dbMutex      = &sync.Mutex{}
	dbOperations = 0
)

type Data interface {
	Key() string
}

type Mapper func(data Data)

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Data (
			Key VARCHAR(200) PRIMARY KEY,
			Value LONGBLOB
		)
	`)
	if err != nil {
		return err
	}

	tx, err = db.Begin()
	if err != nil {
		return err
	}

	return nil
}

func closeDB() error {
	commitDb()
	db.Close()
	return nil
}

func GetData(data Data) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	stmt, err := tx.Prepare(`SELECT Value FROM Data WHERE Key = ?`)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := stmt.Query(data.Key())
	if err != nil {
		log.Fatal(err)
	}

	var rawData []byte
	for rows.Next() {
		if rawData != nil {
			log.Fatalf("more than one result for key: %s", data.Key())
		}
		if err := rows.Scan(&rawData); err != nil {
			log.Fatal(err)
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	if rawData == nil {
		log.Fatalf("no rows found for key: %s", data.Key())
	}

	buf := bytes.NewBuffer(rawData)
	if err := gob.NewDecoder(buf).Decode(data); err != nil {
		log.Fatal(err)
	}
}

func SetData(data Data) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	stmt, err := tx.Prepare(`INSERT INTO Data VALUES (?, ?)`)
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		log.Fatal(err)
	}

	if _, err := stmt.Exec(data.Key(), buf.Bytes()); err != nil {
		log.Fatal(err)
	}

	dbOperations++
	if dbOperations >= config.BufferedOperations {
		commitDb()
	}
}

func MapData(f Mapper) {

}

// Save all the pending transactional data
// Should be called when the dbMutex is hold by this goroutine
func commitDb() {
	if dbOperations > 0 {
		if err := tx.Commit(); err != nil {
			log.Fatal(err)
		}
		dbOperations = 0
	}
}
