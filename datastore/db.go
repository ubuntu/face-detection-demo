package datastore

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

// Stat is a datapoint in time of collected face detected stats
type Stat struct {
	TimeStamp  time.Time
	NumPersons int
}

// Database is the global DB handler
type Database struct {
	dbconn *sql.DB
	// Stats has the full collection of added Stat from the DB or live
	Stats   []Stat
	newstat chan Stat
}

var (
	// DB main object
	DB Database
)

const storagefilename = "storage.db"

// StartDB opens and run the DB in its own goroutine
func StartDB(dir string, shutdown <-chan interface{}, wg *sync.WaitGroup) {
	dbconn, err := sql.Open("sqlite3", path.Join(dir, storagefilename))
	if err != nil {
		log.Fatal("Couldn't open DB", err)
	}

	createTable(dbconn)
	stats, err := fetchAllStats(dbconn)
	if err != nil {
		log.Fatal("Couldn't load DB data", err)
		dbconn.Close()
	}

	DB = Database{dbconn: dbconn, Stats: stats, newstat: make(chan Stat)}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer DB.dbconn.Close()
		defer fmt.Println("Close database")

		for {
			select {
			case s := <-DB.newstat:
				DB.Stats = append(DB.Stats, s)
				DB.insertStat(s)

			case <-shutdown:
				return
			}
		}

	}()

}

func createTable(db *sql.DB) {
	// create table if doesn't exist
	createquery := `
	CREATE TABLE IF NOT EXISTS stats(
		TimeStamp DATETIME,
		NumPersons INTEGER
	);
	`

	if _, err := db.Exec(createquery); err != nil {
		log.Fatal("Couldn't create table", err)
	}
}

func fetchAllStats(db *sql.DB) (result []Stat, err error) {
	readallquery := `
	SELECT TimeStamp, NumPersons FROM stats
	ORDER BY TimeStamp ASC
	`

	rows, err := db.Query(readallquery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		s := Stat{}
		if err = rows.Scan(&s.TimeStamp, &s.NumPersons); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

// Add current stat to the DB and global stats
func (db *Database) Add(s Stat) {
	db.newstat <- s
}

func (db *Database) insertStat(s Stat) {
	addquery := `
	INSERT INTO stats(
		TimeStamp,
		NumPersons
	) values(?, ?)
	`

	stmt, err := db.dbconn.Prepare(addquery)
	if err != nil {
		fmt.Println("Couldn't prepare insert query", err)
		return
	}
	defer stmt.Close()

	if _, err2 := stmt.Exec(s.TimeStamp, s.NumPersons); err2 != nil {
		fmt.Println("Couldn't save", s, ":", err)
	}
}

// WipeDB removes database in dir unconditionally (existing or not)
func WipeDB(dir string) {
	os.Remove(path.Join(dir, storagefilename))
}
