package gatomemes

import (
	"database/sql"
	"errors"
	"fmt"
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

func getRandomLines() (lines [2]string, err error) {
	rows, err := db.Query(
		`SELECT line1, line2 FROM gatomemes AS Q1 JOIN
			(SELECT (RAND() *
				(SELECT MAX(id) FROM gatomemes)) AS id) AS Q2
		WHERE Q1.id >= Q2.id
		ORDER BY Q1.id ASC
		LIMIT 1`,
	)
	if err != nil {
		return lines, err
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&lines[0], &lines[1])
	if err != nil {
		return lines, err
	}

	return lines, nil
}

// FIXME: fails if there is no such id in DB
func getChaoticLines() (lines [2]string, err error) {
	q1, err := getRandomLines()
	if err != nil {
		return lines, err
	}

	q2, err := getRandomLines()
	if err != nil {
		return lines, err
	}

	rand.Seed(time.Now().UnixNano())
	switch rand.Intn(4) {
	case 0:
		lines[0] = q1[0]
		lines[1] = q2[1]
	case 1:
		lines[0] = q2[0]
		lines[1] = q1[1]
	case 2:
		lines[0] = q1[1]
		lines[1] = q2[0]
	case 3:
		lines[0] = q2[1]
		lines[1] = q1[0]
	}
	return lines, nil
}

func addNewUser(login string, password string, identity string) (string, string, error) {
	if identity == "" {
		identity = getUUIDString()
		logger.Println("no identity")
	}

	rows, err := db.Query("SELECT identity FROM user WHERE identity = ?", identity)
	if err != nil {
		logger.Println(err)
	}

	// force generating new identity in case user deleted cookie or id already in DB
	if rows.Next() {
		logger.Println("identity already exists")
		identity = getUUIDString()
	}
	rows.Close()

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		logger.Println("registration was not succesfull", err)
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
		logger.Println("registration was not succesfull", err)
		return "", "", nameErr
	}
	logger.Println("succesfull registration")
	return sessionKey, identity, nil
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
		logger.Println("different identity in DB")
	}

	// generate new session key
	sessionKey = getUUIDString()
	bcryptErr := bcrypt.CompareHashAndPassword([]byte(wantPassword), []byte(gotPassword))
	if bcryptErr == nil {
		logger.Println("successfull login")
		_, err = db.Exec("UPDATE user SET session_key = ? WHERE identity = ?", sessionKey, identityDB)
		if err != nil {
			return "", "", err
		}
		return sessionKey, identityDB, nil
	}

	logger.Println(bcryptErr)
	logger.Println("wrong password")
	return "", "", accessErr
}

func retrieveUserInfo(sessionKey string) (map[string]interface{}, error) {
	row := db.QueryRow("SELECT user_name, reg_time, is_disabled, is_admin, is_root FROM user WHERE session_key = ?", sessionKey)

	var isDisabled, isAdmin, isRoot bool
	var name, regTime string
	err := row.Scan(&name, &regTime, &isDisabled, &isAdmin, &isRoot)
	if err != nil {
		logger.Println("user with given key not found")
		return nil, err
	}

	result := map[string]interface{}{
		"username":   name,
		"regtime":    regTime,
		"isdisabled": isDisabled,
		"isadmin":    isAdmin,
		"isroot":     isRoot,
		"loginerror": "hidden",
		"loginform":  "hidden",
	}
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
	if err != nil {
		logger.Fatal("database init: ", err)
	}

	err = db.Ping()
	if err != nil {
		logger.Fatal("pingErr: ", err)
	}

	logger.Println("succesfully connected to mysql")
}
