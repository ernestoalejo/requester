package requester

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db           *sql.DB
	tx           *sql.Tx
	dbMutex      = &sync.Mutex{}
	dbOperations = 0
)

type Mapper func(key string, data interface{}) error

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return Error(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Data (
			Key VARCHAR(200) PRIMARY KEY,
			Value LONGBLOB
		)
	`)
	if err != nil {
		return Error(err)
	}

	tx, err = db.Begin()
	if err != nil {
		return Error(err)
	}

	return nil
}

func closeDB() error {
	if err := commitDb(true); err != nil {
		return err
	}

	db.Close()
	return nil
}

func GetData(key string, data interface{}) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	stmt, err := tx.Prepare(`SELECT Value FROM Data WHERE Key = ?`)
	if err != nil {
		return Error(err)
	}

	rows, err := stmt.Query(key)
	if err != nil {
		return Error(err)
	}

	var serialized []byte
	for rows.Next() {
		if serialized != nil {
			return Errorf("more than one result for key: %s", key)
		}
		if err := rows.Scan(&serialized); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return Error(err)
	}
	if serialized == nil {
		return Errorf("no rows found for key: %s", key)
	}

	buf := bytes.NewBuffer(serialized)
	if err := gob.NewDecoder(buf).Decode(data); err != nil {
		return Error(err)
	}

	return nil
}

func SetData(key string, data interface{}) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// TODO: Use this idiom for repeated keys
	// INSERT OR IGNORE INTO visits VALUES ($ip, 0);
	// UPDATE visits SET hits = hits + 1 WHERE ip LIKE $ip;
	stmt, err := tx.Prepare(`INSERT INTO Data VALUES (?, ?)`)
	if err != nil {
		return Error(err)
	}

	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		return Error(err)
	}

	if _, err := stmt.Exec(key, buf.Bytes()); err != nil {
		return Error(err)
	}

	dbOperations++
	if err := commitDb(false); err != nil {
		return Error(err)
	}
	return nil
}

func MapData(f Mapper, placeholder interface{}) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := tx.Query(`SELECT Key, Value FROM Data`)
	if err != nil {
		return Error(err)
	}

	for rows.Next() {
		var key string
		var serialized []byte
		if err := rows.Scan(&key, &serialized); err != nil {
			return err
		}

		buf := bytes.NewBuffer(serialized)
		if err := gob.NewDecoder(buf).Decode(placeholder); err != nil {
			return Error(err)
		}

		if err := f(key, placeholder); err != nil {
			return Error(err)
		}
	}
	if err := rows.Err(); err != nil {
		return Error(err)
	}

	return nil
}

// Save all the pending transactional data
// Should be called when the dbMutex is hold by this goroutine
// If the commit is forced, the number of minimum operations before saving
// will not be checked
func commitDb(force bool) error {
	if (force && dbOperations > 0) || dbOperations > config.BufferedOperations {
		actionsLogger.Printf("Commiting %d results...\n", dbOperations)

		if err := tx.Commit(); err != nil {
			return Error(err)
		}
		dbOperations = 0

		var err error
		tx, err = db.Begin()
		if err != nil {
			return Error(err)
		}

		actionsLogger.Printf("Done commiting!")
	}
	return nil
}
