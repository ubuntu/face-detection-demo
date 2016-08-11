package datastore

import (
	"database/sql"
	"fmt"
	"log"
	"path"
)

// Stat is a datapoint in time of collected face detected stats
type Stat struct {
	Time       int
	NumPersons int
}

var (
	db *sql.DB
	// Stats has the full collection of added Stat from the DB or live
	Stats []Stat
)

// LoadDB opens and prepare the db for functioning
func LoadDB(dir string, shutdown <-chan interface{}) {
	var err error
	if db, err = sql.Open("sqlite3", path.Join(dir, "storage.db")); err != nil {
		log.Fatal("Couldn't open DB", err)
	}
	// ensure we close the db before quitting
	go func() {
		<-shutdown
		db.Close()
	}()

	createTable(db)
	Stats, err = fetchAllStats(db)
	if err != nil {
		log.Fatal("Couldn't load DB data", err)
	}
}

// Add current stat to the DB and global stats
func (s Stat) Add() {
	Stats = append(Stats, s)
	go insertStat(db, s)
}

func createTable(db *sql.DB) {
	// create table if doesn't exist
	createquery := `
	CREATE TABLE IF NOT EXISTS stats(
		Time INTEGER,
		NumPersons INTEGER
	);
	`

	if _, err := db.Exec(createquery); err != nil {
		log.Fatal("Couldn't create table", err)
	}
}

func insertStat(db *sql.DB, s Stat) {
	addquery := `
	INSERT INTO stats(
		Time,
		NumPersons
	) values(?, ?)
	`

	stmt, err := db.Prepare(addquery)
	if err != nil {
		fmt.Println("Couldn't prepare insert query", err)
		return
	}
	defer stmt.Close()

	if _, err2 := stmt.Exec(s.Time, s.NumPersons); err2 != nil {
		fmt.Println("Couldn't save", s, ":", err)
	}
}

func fetchAllStats(db *sql.DB) (result []Stat, err error) {
	readallquery := `
	SELECT Time, NumPersons FROM stats
	ORDER BY datetime(Time) DESC
	`

	rows, err := db.Query(readallquery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		s := Stat{}
		if err = rows.Scan(&s.Time, &s.NumPersons); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}
