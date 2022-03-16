package gatomemes

import (
	"log"
	"net/http"
	"os"
)

func checkError(text string, err error) {
	if err != nil {
		log.Fatal(text, err)
	}
}

func GetImageBytes() []byte {
	return imgbytes
}

func GetNew(chaos bool) {
	// get image from web
	if chaos {
		text.dbAccessFunc = getChaoticLines
	} else {
		text.dbAccessFunc = getRandomLines
	}
	resp, err := http.Get(os.Getenv("PROJECTURL"))
	// TODO: return errors to caller
	if err != nil {
		log.Println("response: ", err)
		return
	}
	defer resp.Body.Close()
	dst, err := decodeImage(resp.Header.Get("content-type"), resp.Body)
	if err != nil {
		log.Println("image decoder: ", err)
		return
	}
	fitTextOnImage(dst)
}

func HandleLogin(request *http.Request, identity string) (string, string, error) {
	err := request.ParseForm()
	checkError("HandleLogin: ", err)
	var sessionKey string
	if _, ok := request.PostForm["newuser"]; ok {
		sessionKey, identity, err = addNewUser(request.PostForm["login"][0], request.PostForm["password"][0], identity)
	} else if _, ok := request.PostForm["loginuser"]; ok {
		sessionKey, identity, err = updateSession(request.PostForm["login"][0], request.PostForm["password"][0], identity)
	}
	if sessionKey != "" {
		return sessionKey, identity, nil
	}
	return "", "", err
}

func GetUserInfo(sessionKey string) (result map[string]interface{}, err error) {
	result, err = retrieveUserInfo(sessionKey)
	if err != nil {
		log.Println(err)
		return result, err
	}
	result["loginerror"] = "hidden"
	result["loginform"] = "hidden"
	return result, err
}

func LogOff(sessionKey string) {
	err := deleteSessionKey(sessionKey)
	if err != nil {
		log.Println(err)
	}
}

func GenerateUUID() string {
	return getUUIDString()
}

func init() {
	GetNew(false)
}
