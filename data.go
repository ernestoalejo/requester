package requester

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: DB mutex for the use of the tx.
// TODO: Count the number of operations pending, commit when needed
// TODO: Accept hints of when to save to disk (from the existing process)
// TODO: Get Data.
// TODO: List Data.

var (
	db *sql.DB
	tx *sql.Tx
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
	if err := tx.Commit(); err != nil {
		return err
	}

	db.Close()
	return nil
}

func GetData() Data {
	return nil
}

func SetData(data Data) {
	stmt, err := tx.Prepare(`INSERT INTO Data VALUES (?, ?)`)
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		return err
	}

	if _, err := stmt.Exec(data.Key(), buf.Bytes()); err != nil {
		log.Fatal(err)
	}
}

func MapData(f Mapper) {

}
