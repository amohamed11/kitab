package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

type NoteJSON struct {
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
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

	router.POST("/note", func(c *gin.Context) {
		var json NoteJSON
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var tags []Tag
		for _, s := range json.Tags {
			tag := Tag{Name: s}
			tags = append(tags, tag)
			db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tag)
		}
		note := Note{Content: json.Content, Tags: tags}
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
