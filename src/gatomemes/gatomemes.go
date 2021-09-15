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
		memeText.dbAccessFunc = getChaoticLines
	} else {
		memeText.dbAccessFunc = getRandomLines
	}
	resp, err := http.Get(os.Getenv("PROJECTURL"))
	checkError("response: ", err)
	defer resp.Body.Close()
	convertResponse(resp.Body)
}

func HandleLogin(request *http.Request) (string, error) {
	err := request.ParseForm()
	checkError("HandleLogin: ", err)
	var sessionKey string
	if _, ok := request.PostForm["newuser"]; ok {
		sessionKey, err = addNewUser(request.PostForm["login"][0], request.PostForm["password"][0])
	} else if _, ok := request.PostForm["loginuser"]; ok {
		sessionKey, err = updateSession(request.PostForm["login"][0], request.PostForm["password"][0])
	}
	if sessionKey != "" {
		return sessionKey, nil
	}
	return "", err
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

func init() {
	GetNew(false)
}
