package gatomemes

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
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

func GetImage(key string) ([]byte, error) {
	data, err := rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetRandomImageID() (string, error) {
	id, err := rdb.SRandMember(context.Background(), "results").Result()
	if err == redis.Nil {
		return CreateNew(false)
	} else if err != nil {
		return "", err
	}

	valid, err := IsValidID(id)
	if err != nil {
		return "", err
	}

	if !valid {
		logger.Println("image id:", id, "is gone")
		rdb.SRem(context.Background(), "results", id)
		return GetRandomImageID()
	}

	return id, nil
}

func IsValidID(id string) (bool, error) {
	result, err := rdb.Exists(context.Background(), id).Result()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func isValidURL(link string) bool {
	u, err := url.Parse(link)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func CreateNew(chaos bool) (string, error) {
	var lines [2]string
	var err error

	if chaos {
		lines, err = getChaoticLines()
	} else {
		lines, err = getRandomLines()
	}

	if err != nil {
		return "", err
	}

	// FIXME: input string is very hacky
	img, err := CreateNewFromSrc(os.Getenv("APP_SOURCE"), lines[0]+"@@"+lines[1], nil)
	if err != nil {
		return "", err
	}

	h := sha1.New()
	h.Write(img)
	id := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	err = rdb.Set(context.Background(), id, img, time.Minute).Err()
	if err != nil {
		return "", err
	}

	err = rdb.SAdd(context.Background(), "results", id).Err()
	if err != nil {
		return "", err
	}

	return id, nil
}

// FIXME: very poorly organised, base64 input is not checked until read fully
func handleURL(link string) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "GET", link, nil)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// FIXME: dirty fix for sites like thiscatdoesnotexist.com, where
	// one url leads to different images with different ETags,
	// append ETag to url before checking in cache
	link = fmt.Sprintf("%s%s", link, resp.Header.Get("ETag"))

	var mimeType string
	data, err := rdb.Get(context.Background(), link).Bytes()
	if err == redis.Nil {
		mimeType = resp.Header.Get("content-type")
		// do not even attempt to download wrong content type data
		if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/bmp" {
			return nil, "", errors.New("unsupported data type")
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}

		err = rdb.Set(context.Background(), link, data, 24*time.Hour).Err()
		if err != nil {
			return nil, "", err
		}

	} else if err != nil {
		return nil, "", err

	} else {
		logger.Println("match:", link)
	}

	return data, mimeType, nil
}

func CreateNewFromSrc(src string, text string, opt *Options) ([]byte, error) {
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
	} else if data, err = base64.StdEncoding.DecodeString(src); err != nil {
		return nil, errors.New("source unrecognized")
	}

	dst, err := decodeImage(data, dataType)
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

	img, err := encodePNG(dst)
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
		Addr:     os.Getenv("RDB_HOST"),
		Username: os.Getenv("RDB_USER"),
		Password: os.Getenv("RDB_PASS"),
		DB:       0,
	})
}
