package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/olde-ducke/gatomemes/src/gatomemes"
)

// rendering template, template paramers unused for now
func rootHandler(context *gin.Context) {
	context.HTML(http.StatusOK, "index.html", gin.H{"image": "getNewGatito()"})
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

func main() {
	// server
	router := gin.Default()

	router.LoadHTMLFiles("templates/index.html")
	router.GET("/", rootHandler)
	router.GET("/gato.png", imageHandler)
	router.GET("/new", newHandler)
	router.GET("/chaos", chaosHandler)
	router.GET("/test", testHandler)
	router.Run(":8080")
}
