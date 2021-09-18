package gatomemes

import (
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var db *sql.DB

func getUUIDString() string {
	return uuid.New().String()
}

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

func addNewUser(login string, password string, identity string) (string, string, error) {
	log.Println("addNewUser: ", login, password)
	rows, err := db.Query("SELECT identity FROM user WHERE EXISTS (SELECT identity FROM user WHERE identity = ?)", identity)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	if rows.Next() || identity == "" {
		log.Println("identity already exists or cookie deleted")
		identity = getUUIDString()
	}

	sessionKey := getUUIDString()
	nameErr := errors.New("name_taken")
	// TODO: for now password is stored as plaintext
	_, err = db.Exec("INSERT INTO user (identity, user_name, password, session_key) VALUES (?, ?, ?, ?)",
		identity, login, password, sessionKey)
	if err != nil {
		log.Println("registration was not succesfull", err)
		return "", "", nameErr
	} else {
		log.Println("succesfull registration")
		return sessionKey, identity, nil
	}
}

func updateSession(login string, gotPassword string, identity string) (sessionKey string, identityDB string, accessErr error) {
	log.Println("updateSession: ", login, gotPassword)
	// TODO: handle internal errors
	rows, err := db.Query("SELECT identity, password FROM user WHERE user_name = ?", login)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	accessErr = errors.New("wrong_credentials")
	if !rows.Next() {
		log.Println("wrong login")
		return "", "", accessErr
	}

	var wantPassword string
	// generate new session key
	sessionKey = getUUIDString()

	err = rows.Scan(&identityDB, &wantPassword)
	checkError("loginUser", err)

	// FIXME: string comparison is bad?
	if identity != identityDB {
		log.Println("different identity in DB")
	}

	if gotPassword == wantPassword {
		log.Println("successfull login")
		_, err = db.Exec("UPDATE user SET session_key = ? WHERE identity = ?", sessionKey, identityDB)
		checkError("updateSession", err)
		if err == nil {
			return sessionKey, identityDB, nil
		}
	}
	log.Println("wrong password")
	return "", "", accessErr
}

func retrieveUserInfo(sessionKey string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	rows, err := db.Query("SELECT identity, user_name, password, reg_time, is_disabled FROM user WHERE session_key = ?", sessionKey)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	if !rows.Next() {
		err = errors.New("user with given key not found")
		return result, err
	}
	var identity string
	var is_disabled bool
	var name, password, regTime string
	err = rows.Scan(&identity, &name, &password, &regTime, &is_disabled)
	if err != nil {
		return result, err
	}
	result["identity"] = identity
	result["username"] = name
	result["password"] = password
	result["regtime"] = regTime
	result["isdisabled"] = is_disabled
	return result, nil
}

func deleteSessionKey(sessionKey string) (err error) {
	_, err = db.Exec("UPDATE user SET session_key = NULL WHERE session_key = ?", sessionKey)
	return err
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
