package server

import (
	"net/http"

	"git.sr.ht/~anecdotal/kitab/notes"
	"github.com/gin-gonic/gin"
)

func Init(port string) {
	// Setup router
	router := gin.Default()
	router.SetTrustedProxies([]string{"localhost"})
	router.LoadHTMLGlob("templates/**/*.tmpl")
	router.Static("/assets", "./assets")

	notesController := notes.NoteController{}
	notesController.Init()

	router.GET("/", notesController.Index)
	router.GET("/notes/:id", notesController.GetById)
	router.GET("/notes/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "notes/new.tmpl", gin.H{})
	})
	router.POST("/notes/new", notesController.New)
	router.POST("/notes/search", notesController.Search)

	// Listen and serve on 0.0.0.0:8080
	router.Run(":8080")
}
