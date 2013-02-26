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

	selectStmt, insertStmt, updateStmt *sql.Stmt
)

type Mapper func(key string, data interface{}) error
type Creator func() interface{}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "file:./data.db?cache=shared&mode=rwc")
	if err != nil {
		return Error(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Data (
			Key VARCHAR(2000) PRIMARY KEY,
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

	selectStmt, err = db.Prepare(`SELECT Value FROM Data WHERE Key = ?`)
	if err != nil {
		return Error(err)
	}

	if err := prepareTransactionalStmt(); err != nil {
		return err
	}

	return nil
}

func prepareTransactionalStmt() error {
	var err error

	insertStmt, err = tx.Prepare(`INSERT OR IGNORE INTO Data VALUES (?, NULL)`)
	if err != nil {
		return Error(err)
	}
	updateStmt, err = tx.Prepare(`UPDATE Data SET Value = ? WHERE Key = ?`)
	if err != nil {
		return Error(err)
	}

	return nil
}

func closeDB() error {
	if err := commitDb(true); err != nil {
		return err
	}

	selectStmt.Close()
	insertStmt.Close()
	db.Close()
	return nil
}

func GetData(key string, data interface{}) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if err := commitDb(true); err != nil {
		return err
	}

	rows, err := selectStmt.Query(key)
	if err != nil {
		return Error(err)
	}
	defer rows.Close()
	var serialized []byte
	for rows.Next() {
		if serialized != nil {
			return Errorf("more than one result for key: %s", key)
		}
		if err := rows.Scan(&serialized); err != nil {
			return Error(err)
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

func CountData() (int, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if err := commitDb(true); err != nil {
		return 0, err
	}

	rows, err := db.Query(`SELECT COUNT(*) FROM Data;`)
	if err != nil {
		return 0, Error(err)
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, Error(err)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, Error(err)
	}

	return count, nil
}

func SetData(key string, data interface{}) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if _, err := insertStmt.Exec(key); err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		return Error(err)
	}
	if _, err := updateStmt.Exec(buf.Bytes(), key); err != nil {
		return Error(err)
	}

	dbOperations++
	if err := commitDb(false); err != nil {
		return Error(err)
	}
	return nil
}

func MapData(f Mapper, creator Creator) error {
	return MapDataLimit(f, creator, -1)
}

func MapDataLimit(f Mapper, creator Creator, limit int) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query(`SELECT Key, Value FROM Data`)
	if err != nil {
		return Error(err)
	}

	count := 0
	for rows.Next() {
		var key string
		var serialized []byte
		if err := rows.Scan(&key, &serialized); err != nil {
			return Error(err)
		}

		data := creator()
		buf := bytes.NewBuffer(serialized)
		if err := gob.NewDecoder(buf).Decode(data); err != nil {
			return Error(err)
		}

		if err := f(key, data); err != nil {
			return Error(err)
		}

		count++
		if limit > -1 && count >= limit {
			break
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

		if err := prepareTransactionalStmt(); err != nil {
			return err
		}

		actionsLogger.Printf("Done commiting!")
	}
	return nil
}
