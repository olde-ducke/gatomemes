package gatomemes

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	rows, err := db.Query("SELECT identity FROM user WHERE identity = ?", identity)
	if err != nil {
		log.Println(err)
	}
	// force generating new identity in case user deleted cookie or id already in DB
	if rows.Next() || identity == "" {
		log.Println("identity already exists or cookie deleted")
		identity = getUUIDString()
	}
	rows.Close()

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Println("registration was not succesfull", err)
		return "", "", err
	}
	// FIXME: is this really necessary?
	var sbuilder strings.Builder
	for _, v := range encryptedPassword {
		fmt.Fprintf(&sbuilder, "%c", v)
	}

	sessionKey := getUUIDString()
	nameErr := errors.New("name_taken")
	_, err = db.Exec("INSERT INTO user (identity, user_name, password, session_key) VALUES (?, ?, ?, ?)",
		identity, login, sbuilder.String(), sessionKey)
	if err != nil {
		log.Println("registration was not succesfull", err)
		return "", "", nameErr
	} else {
		log.Println("succesfull registration")
		return sessionKey, identity, nil
	}
}

func updateSession(login string, gotPassword string, identity string) (sessionKey string, identityDB string, accessErr error) {
	// TODO: handle internal errors
	row := db.QueryRow("SELECT identity, password FROM user WHERE user_name = ?", login)
	var wantPassword string
	err := row.Scan(&identityDB, &wantPassword)
	accessErr = errors.New("wrong_credentials")
	if err != nil {
		return "", "", accessErr
	}

	if identity != identityDB {
		log.Println("different identity in DB")
	}

	// generate new session key
	sessionKey = getUUIDString()
	if bcrypterr := bcrypt.CompareHashAndPassword([]byte(wantPassword), []byte(gotPassword)); bcrypterr == nil {
		log.Println("successfull login")
		_, err = db.Exec("UPDATE user SET session_key = ? WHERE identity = ?", sessionKey, identityDB)
		// FIXME: checkerror fatals on errors
		checkError("updateSession", err)
		if err == nil {
			return sessionKey, identityDB, nil
		}
	} else {
		log.Println(bcrypterr)
	}
	log.Println("wrong password")
	return "", "", accessErr
}

func retrieveUserInfo(sessionKey string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	row := db.QueryRow("SELECT user_name, reg_time, is_disabled, is_admin, is_root FROM user WHERE session_key = ?", sessionKey)
	var isDisabled, isAdmin, isRoot bool
	var name, regTime string
	err = row.Scan(&name, &regTime, &isDisabled, &isAdmin, &isRoot)
	if err != nil {
		log.Println("user with given key not found")
		return result, err
	}
	result["username"] = name
	result["regtime"] = regTime
	result["isdisabled"] = isDisabled
	result["isadmin"] = isAdmin
	result["isroot"] = isRoot
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
