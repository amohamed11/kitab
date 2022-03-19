package main

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Note struct {
	gorm.Model
	Title   string
	Content string
	Tags    []Tag `gorm:"many2many:note_tags;"`
}

type Tag struct {
	gorm.Model
	Name string
}

type NoteForm struct {
	Title   string `form:"title"`
	Content string `form:"content"`
	Tags    string `form:"tags"`
}

func main() {
	// Initialize database connection
	db, err := gorm.Open(sqlite.Open("kitab.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&Note{})

	// Setup goldmark
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	// Setup server
	router := gin.Default()
	router.SetTrustedProxies([]string{"localhost"})
	router.LoadHTMLGlob("templates/**/*.tmpl")
	router.Static("/assets", "./assets")

	router.GET("/", func(c *gin.Context) {
		var notes []Note
		result := db.Find(&notes)

		if result.Error == nil {
			c.HTML(http.StatusOK, "home/index.tmpl", gin.H{
				"count": len(notes),
				"notes": notes,
			})
			return
		}
		c.HTML(http.StatusInternalServerError, "shared/error.tmpl", gin.H{
			"msg":   "Something went wrong!",
			"error": result.Error,
		})
	})

	router.GET("/notes/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "notes/new.tmpl", gin.H{})
	})

	router.POST("/notes/new", func(c *gin.Context) {
		var form NoteForm
		if err := c.ShouldBind(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var tags []Tag
		splitTags := strings.Split(form.Tags, " ")
		for _, s := range splitTags {
			var tag Tag
			db.FirstOrCreate(&tag, Tag{Name: s})
			tags = append(tags, tag)
		}
		note := Note{Title: form.Title, Content: form.Content, Tags: tags}
		result := db.Create(&note)

		if result.RowsAffected > 0 && result.Error == nil {
			var buf bytes.Buffer
			if err := md.Convert([]byte(note.Content), &buf); err != nil {
				panic(err)
			}
			c.HTML(http.StatusCreated, "notes/view.tmpl", gin.H{
				"note":        note,
				"contentHTML": template.HTML(buf.String()),
			})
			return
		}

		c.HTML(http.StatusInternalServerError, "shared/error.tmpl", gin.H{
			"msg":   "Something went wrong!",
			"error": result.Error,
		})
	})

	router.GET("/notes/:id", func(c *gin.Context) {
		var note Note
		id := c.Param("id")
		result := db.Preload("Tags").Find(&note, id)

		if result.Error == nil {
			var buf bytes.Buffer
			if err := md.Convert([]byte(note.Content), &buf); err != nil {
				panic(err)
			}
			c.HTML(http.StatusOK, "notes/view.tmpl", gin.H{
				"note":        note,
				"contentHTML": template.HTML(buf.String()),
			})
			return
		}
		c.HTML(http.StatusInternalServerError, "shared/error.tmpl", gin.H{
			"msg":   "Something went wrong!",
			"error": result.Error,
		})
	})

	// Listen and serve on 0.0.0.0:8080
	router.Run(":8080")
}
