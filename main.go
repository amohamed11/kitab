package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Note struct {
	gorm.Model
	Content string
	Tags    []Tag `gorm:"many2many:note_tags;"`
}

type Tag struct {
	gorm.Model
	Name string
}

type NoteForm struct {
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

	// Setup server
	router := gin.Default()
	router.SetTrustedProxies([]string{"localhost"})
	router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./assets")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Kitab",
		})
	})

	router.POST("/note", func(c *gin.Context) {
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
		note := Note{Content: form.Content, Tags: tags}
		result := db.Create(&note)

		if result.RowsAffected > 0 {
			c.JSON(http.StatusCreated, gin.H{"body": note})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": result.Error})
	})

	router.GET("/note", func(c *gin.Context) {
		var notes []Note
		result := db.Find(&notes)

		if result.RowsAffected > 1 {
			c.JSON(http.StatusOK, gin.H{"body": notes})
		}

		c.JSON(http.StatusOK, gin.H{"body": "No notes created yet!"})
	})

	// Listen and serve on 0.0.0.0:8080
	router.Run(":8080")
}
