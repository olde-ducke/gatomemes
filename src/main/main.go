package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/olde-ducke/gatomemes/src/gatomemes"
)

func rootHandler(c *gin.Context) {
	id, err := gatomemes.GetRandomImageID()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	if c.IsAborted() {
		errorHandler(c)
		return
	}

	c.Redirect(http.StatusFound, "/page/"+id)
}

// rendering template
func pageHandler(c *gin.Context) {
	id := c.Param("id")
	valid, err := gatomemes.IsValidID(id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	if !valid {
		c.AbortWithStatus(http.StatusNotFound)
	}

	if c.IsAborted() {
		errorHandler(c)
		return
	}

	getIdentity(c)
	text, err := c.Cookie("error")
	// TODO: server internal errors
	// if there is an error cookie change template accordingly
	if err == nil {
		log.Println("no session cookie")
		c.SetCookie("error", "", -1, "/", "", false, true)
		if text == "wrong_credentials" {
			text = "nombre de usuario/contraseña incorrectos"
		} else {
			text = "se toma el nombre de usuario "
		}
		c.HTML(http.StatusUnauthorized, "index.html", gin.H{
			"errortext": text,
			"userinfo":  "hidden",
		})
		return
	}

	sessionKey, err := c.Cookie("sessionkey")
	if err != nil {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"id":         id,
			"loginerror": "hidden",
			"userinfo":   "hidden",
		})
		return
	}

	result, err := gatomemes.GetUserInfo(sessionKey)
	if err != nil {
		c.SetCookie("sessionkey", "", -1, "/", "", false, true) // ???
		c.HTML(http.StatusOK, "index.html", gin.H{
			"id":         id,
			"loginerror": "hidden",
			"userinfo":   "hidden",
		})
		return
	}
	result["id"] = id
	c.HTML(http.StatusOK, "index.html", result)
}

func imageHandler(c *gin.Context) {
	id := strings.TrimSuffix(c.Param("id"), ".png")
	imgBytes, err := gatomemes.GetImage(id)
	if err == redis.Nil {
		c.AbortWithStatus(http.StatusNotFound)
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	if c.IsAborted() {
		errorHandler(c)
		return
	}

	c.Data(http.StatusOK, "image/png", imgBytes)
}

func newHandler(c *gin.Context) {
	id, err := gatomemes.CreateNew(c.Param("handler") == "chaotic")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		errorHandler(c)
		return
	}
	c.Redirect(http.StatusFound, "/page/"+id)
}

func loginFormHandler(c *gin.Context) {
	sessionKey, identity, err := gatomemes.HandleLogin(c.Request, getIdentity(c))
	if err != nil {
		c.SetCookie("error", err.Error(), 86400, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	} else {
		c.SetCookie("sessionkey", sessionKey, 86400, "/", "", false, true)
		c.SetCookie("identity", identity, 86400, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	}
}

func errorHandler(c *gin.Context) {
	status := c.Writer.Status()
	c.Writer.WriteString(strconv.Itoa(status) + " " + http.StatusText(status))
}

func logoutHandler(c *gin.Context) {
	sessionKey, err := c.Cookie("sessionkey")
	if err != nil {
		log.Println(err)
		c.Redirect(http.StatusFound, "/")
	}
	gatomemes.LogOff(sessionKey)
	c.SetCookie("sessionkey", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

func getIdentity(c *gin.Context) string {
	c.SetSameSite(http.SameSiteStrictMode)
	identity, err := c.Cookie("identity")
	if err != nil {
		identity = gatomemes.GenerateUUID()
		c.SetCookie("identity", identity, 86400, "/", "", false, true)
	}
	return identity
}

func main() {
	go grpcServerRun()
	router := gin.Default()

	router.LoadHTMLFiles("templates/index.html")
	router.GET("/", rootHandler)
	router.GET("/page/:id", pageHandler)
	router.GET("/gato/:id", imageHandler)
	router.GET("/new/:handler", newHandler)
	router.POST("/login", loginFormHandler)
	router.GET("/logout", logoutHandler)
	router.NoRoute(errorHandler)
	router.Run(":8080")
	// TODO: fix static files
}
