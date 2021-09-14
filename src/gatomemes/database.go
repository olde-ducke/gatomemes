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
	rows, err := db.Query("SELECT line1, line2 FROM gatomemes WHERE id = ?", getRandomId())
	checkError("getRandimLines: ", err)
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&lines[0], &lines[1])
	checkError("getRandimLines: ", err)
	return lines
}

func getChaoticLines() (lines [2]string) {
	rows, err := db.Query("SELECT Q1.line1, Q2.line2 FROM gatomemes Q1, gatomemes Q2 WHERE Q1.id = ? and Q2.id = ?",
		getRandomId(), getRandomId())
	checkError("getChaoticLines: ", err)
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&lines[0], &lines[1])
	checkError("getChaoticLines: ", err)
	return lines
}

func getRandomId() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(getMaxId()) + 1
}

func getMaxId() (id int) {
	rows, err := db.Query("SELECT MAX(id) FROM gatomemes")
	checkError("getMaxId: ", err)
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&id)
	checkError("getMaxId: ", err)
	return id
}

func addNewUser(login string, password string) {
	log.Println("addNewUser: ", login, password)
	// TODO: for now just stores Unix time as session key
	sessionKey := time.Now().UnixMicro()
	// TODO: for now password is stored as plaintext
	// TODO: tell frontend that username is taken
	_, err := db.Exec("INSERT INTO user (user_name, password, session_key) VALUES (?, ?, ?)", login, password, sessionKey)
	if err != nil {
		log.Println("registration was not succesfull", err)
	} else {
		log.Println("succesfull registration")
	}
}

func loginUser(login string, gotPassword string) {
	log.Println("loginUser: ", login, gotPassword)
	// TODO: don't just crash server on the wrong login
	rows, err := db.Query("SELECT id, password FROM user WHERE user_name = ?", login)
	if !rows.Next() {
		log.Println("wrong login")
		return
	}
	defer rows.Close()

	var wantPassword string
	var id int64
	err = rows.Scan(&id, &wantPassword)
	checkError("loginUser", err)

	if gotPassword == wantPassword {
		log.Println("successfull login")
		sessionKey := time.Now().UnixMicro()
		_, err = db.Exec("UPDATE user SET session_key = ? WHERE id = ?", sessionKey, id)
		checkError("addNewUser", err)
	} else {
		log.Println("wrong password")
	}
	// TODO: for now just stores Unix time as session key
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
