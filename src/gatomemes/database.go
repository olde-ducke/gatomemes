package gatomemes

import (
	"database/sql"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func getRandomLines() (lines [2]string) {
	rows, err := db.Query("SELECT Line1, Line2 FROM gatomemes WHERE Id = ?", getRandomId())
	checkError("db.Query: ", err)
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&lines[0], &lines[1])
	checkError("rows.Scan:", err)
	return lines
}

func getChaoticLines() (lines [2]string) {
	rows, err := db.Query("SELECT Q1.Line1, Q2.Line2 FROM gatomemes Q1, gatomemes Q2 WHERE Q1.Id = ? and Q2.Id = ?",
		getRandomId(), getRandomId())
	checkError("db.Query: ", err)
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&lines[0], &lines[1])
	checkError("rows.Scan:", err)
	return lines
}

func getRandomId() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(getMaxId()) + 1
}

func getMaxId() (id int) {
	rows, err := db.Query("SELECT MAX(Id) FROM gatomemes")
	checkError("db.Query: ", err)
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&id)
	checkError("rows.Scan:", err)
	return id
}

func init() {
	// setup DB
	// Capture connection properties.
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "gatomemes",
	}
	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	checkError("database handle: ", err)

	err = db.Ping()
	checkError("pingErr: ", err)
	log.Println("Connected to DB!")
}
