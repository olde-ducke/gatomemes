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

func HandleLogin(request *http.Request) {
	err := request.ParseForm()
	checkError("HandleLogin: ", err)
	if _, ok := request.PostForm["newuser"]; ok {
		addNewUser(request.PostForm["login"][0], request.PostForm["password"][0])
	} else if _, ok := request.PostForm["loginuser"]; ok {
		loginUser(request.PostForm["login"][0], request.PostForm["password"][0])
	}
}

func init() {
	GetNew(false)
}
