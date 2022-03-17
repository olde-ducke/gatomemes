package gatomemes

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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

	var lines [2]string
	if chaos {
		lines, err = getChaoticLines()
	} else {
		lines, err = getRandomLines()
	}
	if err != nil {
		log.Println("request to db failed: ", err)
		return
	}
	drawGlyph(lines[0], &options{outlineWidth: 10.0}, dst, top)
	drawGlyph(lines[1], &options{outlineWidth: 10.0}, dst, bottom)
	encodeImage(dst)
}

func isValidURL(link string) bool {
	u, err := url.Parse(link)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func GetNewFromSRC(src string, text string) (image.Image, error) {
	if src == "" || text == "" {
		return nil, errors.New("image source or text is empty")
	}

	var reader io.Reader
	var dataType string
	var err error

	if isValidURL(src) {
		resp, err := http.Get(src)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		reader = resp.Body
		dataType = resp.Header.Get("content-type")
	} else if data, err := base64.StdEncoding.DecodeString(src); err == nil {
		reader = bytes.NewReader(data)
		dataType = http.DetectContentType(data)
	} else {
		return nil, errors.New("bad input")
	}

	dst, err := decodeImage(dataType, reader)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(text, "\n")
	for alignment, text := range lines {
		if alignment > 2 {
			break
		}
		drawGlyph(text, &options{outlineWidth: 10.0}, dst, alignment)
	}
	encodeImage(dst)
	return dst, nil
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
