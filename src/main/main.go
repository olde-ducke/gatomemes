package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/olde-ducke/gatomemes/src/gatomemes"
)

// rendering template, template paramers unused for now
func rootHandler(context *gin.Context) {
	text, err := context.Cookie("error")
	// TODO: server internal errors
	if err == nil {
		log.Println("error cookie")
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
		log.Println("no cookie")
		return
	}
	//gatomemes.GetUserInfo(sessionKey)
	log.Println(sessionKey)
	result, err := gatomemes.GetUserInfo(sessionKey)
	if err != nil {
		context.HTML(http.StatusInternalServerError, "index.html", gin.H{"loginerror": "hidden", "loginform": "hidden"})
	}
	log.Println(result)
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
	context.Redirect(http.StatusFound, "/")
}

func loginFormHandler(context *gin.Context) {
	context.SetSameSite(http.SameSiteStrictMode)
	sessionKey, err := gatomemes.HandleLogin(context.Request)
	if err != nil {
		context.SetCookie("error", err.Error(), 86400, "/", "localhost", true, true)
		context.Redirect(http.StatusFound, "/")
	} else {
		context.SetCookie("sessionkey", sessionKey, 86400, "/", "localhost", true, true)
		context.Redirect(http.StatusFound, "/")
	}
}

func logoutHandler(context *gin.Context) {
	context.SetCookie("sessionkey", "", -1, "/", "localhost", true, true)
	context.Redirect(http.StatusFound, "/")
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
