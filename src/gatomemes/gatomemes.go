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
	convertRespons(resp.Body)
}

func init() {
	GetNew(false)
}
