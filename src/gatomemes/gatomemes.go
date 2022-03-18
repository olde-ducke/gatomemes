package gatomemes

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client
var logger = log.New(os.Stdout, "\x1b[31m[GAT] \x1b[0m", log.LstdFlags)

func fatalError(text string, err error) {
	if err != nil {
		log.Fatal(text, err)
	}
}

func GetImage(key string) ([]byte, error) {
	data, err := rdb.Get(context.Background(), key).Bytes()
	if err == redis.Nil {
		return GetNew(key, false)
	} else if err != nil {
		return nil, err
	}
	return data, nil
}

func isValidURL(link string) bool {
	u, err := url.Parse(link)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func GetNew(key string, chaos bool) ([]byte, error) {
	var lines [2]string
	var err error
	if chaos {
		lines, err = getChaoticLines()
	} else {
		lines, err = getRandomLines()
	}
	// FIXME: no error checking

	// FIXME: input string is very hacky
	img, err := GetNewFromSrc(os.Getenv("PROJECTURL"), lines[0]+"@@"+lines[1], nil)
	if err != nil {
		return nil, err
	}

	err = rdb.Set(context.Background(), key, img, time.Minute).Err()
	return img, err
}

func handleURL(link string) ([]byte, string, error) {
	var dataType string
	data, err := rdb.Get(context.Background(), link).Bytes()
	if err == redis.Nil {
		resp, err := http.Get(link)
		if err != nil {
			return nil, "", err
		}
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}

		err = rdb.Set(context.Background(), link, data, time.Minute).Err()
		if err != nil {
			return nil, "", err
		}

		dataType = resp.Header.Get("content-type")
	} else if err != nil {
		return nil, "", err
	} else {
		dataType = http.DetectContentType(data)
	}
	return data, dataType, nil
}

func GetNewFromSrc(src string, text string, opt *Options) ([]byte, error) {
	if src == "" {
		return nil, errors.New("image source is empty")
	}

	if text == "" {
		return nil, errors.New("text is empty")
	}

	var data []byte
	var dataType string
	var err error

	if isValidURL(src) {
		data, dataType, err = handleURL(src)
	} else if data, err = base64.StdEncoding.DecodeString(src); err == nil {
		dataType = http.DetectContentType(data)
	} else {
		return nil, errors.New("source unrecognized")
	}

	dst, err := decodeImage(dataType, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	lines := strings.Split(text, "@")
	for vAlignment, text := range lines {
		if vAlignment > 2 {
			break
		}
		drawGlyphs(text, opt, dst, vAlignment)
	}

	img, err := encodeImage(dst)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func HandleLogin(request *http.Request, identity string) (string, string, error) {
	err := request.ParseForm()
	if err != nil {
		return "", "", err
	}

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
		logger.Println(err)
		return result, err
	}
	result["loginerror"] = "hidden"
	result["loginform"] = "hidden"
	return result, err
}

func LogOff(sessionKey string) {
	err := deleteSessionKey(sessionKey)
	if err != nil {
		logger.Println(err)
	}
}

func GenerateUUID() string {
	return getUUIDString()
}

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: os.Getenv("RDBPASS"),
		DB:       0,
	})
}
