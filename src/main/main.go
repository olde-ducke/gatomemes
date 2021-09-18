package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/olde-ducke/gatomemes/src/gatomemes"
)

// rendering template
func rootHandler(context *gin.Context) {
	getIdentity(context)
	text, err := context.Cookie("error")
	// TODO: server internal errors
	// if there is an error cookie change template accordingly
	if err == nil {
		log.Println("no session cookie")
		context.SetCookie("error", "", -1, "/", "localhost", true, true)
		if text == "wrong_credentials" {
			text = "nombre de usuario/contraseña incorrectos"
		} else {
			text = "se toma el nombre de usuario "
		}
		context.HTML(http.StatusUnauthorized, "index.html", gin.H{"errortext": text, "userinfo": "hidden"})
		return
	}

	sessionKey, err := context.Cookie("sessionkey")
	if err != nil {
		context.HTML(http.StatusOK, "index.html", gin.H{"loginerror": "hidden", "userinfo": "hidden"})
		log.Println("sessionkey not found: ", err)
		return
	}
	//gatomemes.GetUserInfo(sessionKey)
	result, err := gatomemes.GetUserInfo(sessionKey)
	if err != nil {
		context.SetCookie("sessionkey", "", -1, "/", "localhost", true, true) // ???
		context.HTML(http.StatusOK, "index.html", gin.H{"loginerror": "hidden", "userinfo": "hidden"})
		return
	}
	//log.Println(result)
	context.HTML(http.StatusOK, "index.html", result)
}

// fake /gato.jpeg response
func imageHandler(context *gin.Context) {
	context.Data(http.StatusOK, "image/png", gatomemes.GetImageBytes())
}

func newHandler(context *gin.Context) {
	gatomemes.GetNew(false)
	context.Redirect(http.StatusFound, "/")
}

func chaosHandler(context *gin.Context) {
	gatomemes.GetNew(true)
	context.Redirect(http.StatusFound, "/")
}

func testHandler(context *gin.Context) {
	gatomemes.DrawTestOutline()
	//context.Redirect(http.StatusFound, "/")
}

func loginFormHandler(context *gin.Context) {
	sessionKey, identity, err := gatomemes.HandleLogin(context.Request, getIdentity(context))
	if err != nil {
		context.SetCookie("error", err.Error(), 86400, "/", "localhost", true, true)
		context.Redirect(http.StatusFound, "/")
	} else {
		context.SetCookie("sessionkey", sessionKey, 86400, "/", "localhost", true, true)
		context.SetCookie("identity", identity, 86400, "/", "localhost", true, true)
		context.Redirect(http.StatusFound, "/")
	}
}

func logoutHandler(context *gin.Context) {
	sessionKey, err := context.Cookie("sessionkey")
	if err != nil {
		log.Println(err)
		context.Redirect(http.StatusFound, "/")
	}
	gatomemes.LogOff(sessionKey)
	context.SetCookie("sessionkey", "", -1, "/", "localhost", true, true)
	context.Redirect(http.StatusFound, "/")
}

func getIdentity(context *gin.Context) string {
	context.SetSameSite(http.SameSiteStrictMode)
	identity, err := context.Cookie("identity")
	if err != nil {
		context.SetCookie("identity", gatomemes.GenerateUUID(), 86400, "/", "localhost", true, true)
	}
	return identity
}

func main() {
	// server
	router := gin.Default()

	router.LoadHTMLFiles("templates/index.html")
	router.GET("/", rootHandler)
	router.GET("/gato.png", imageHandler)
	router.GET("/new", newHandler)
	router.GET("/chaos", chaosHandler)
	router.GET("/test", testHandler)
	router.POST("/login", loginFormHandler)
	router.GET("/logout", logoutHandler)
	router.Run(":8080")
}
