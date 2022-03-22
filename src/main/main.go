package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/olde-ducke/gatomemes/src/gatomemes"
)

// rendering template
func rootHandler(c *gin.Context) {
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
		c.HTML(http.StatusUnauthorized, "index.html", gin.H{"errortext": text, "userinfo": "hidden"})
		return
	}

	sessionKey, err := c.Cookie("sessionkey")
	if err != nil {
		c.HTML(http.StatusOK, "index.html", gin.H{"loginerror": "hidden", "userinfo": "hidden"})
		// log.Println("sessionkey not found: ", err)
		return
	}
	//gatomemes.GetUserInfo(sessionKey)
	result, err := gatomemes.GetUserInfo(sessionKey)
	if err != nil {
		c.SetCookie("sessionkey", "", -1, "/", "", false, true) // ???
		c.HTML(http.StatusOK, "index.html", gin.H{"loginerror": "hidden", "userinfo": "hidden"})
		return
	}
	//log.Println(result)
	c.HTML(http.StatusOK, "index.html", result)
}

// fake /gato.jpeg response
func imageHandler(c *gin.Context) {
	identity := getIdentity(c)
	imgBytes, err := gatomemes.GetImage(identity)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "image/png", imgBytes)
}

func newHandler(c *gin.Context) {
	_, err := gatomemes.GetNew(c.Param("handler") == "chaotic")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusFound, "/")
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
	router.GET("/gato.png", imageHandler)
	router.GET("/new/*handler", newHandler)
	router.POST("/login", loginFormHandler)
	router.GET("/logout", logoutHandler)
	router.Run(":8080")
	// TODO: fix static files
}
