package db

import (
	"database/sql"
	"fmt"
)

var (
	// shared db handle for db package
	paoDb *sql.DB
)

// Init initializes the database with necessary tables
func Init(db *sql.DB) {
	paoDb = db
	_, err := db.Exec(`create table if not exists completedGames(id SERIAL, winner varchar(255), loser varchar(255), winColor varchar(255), primary key(id))`)
	if err != nil {
		fmt.Printf("Could not create db table: %s\n", err)
	}
}
