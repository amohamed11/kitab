package main

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Server struct {
	DB *gorm.DB
	MD goldmark.Markdown
}

type Note struct {
	gorm.Model
	Title   string
	Content string
}

type NoteSearch struct {
	ID        int
	Noterowid int
	Content   string
}

type NoteForm struct {
	Title   string `form:"title"`
	Content string `form:"content"`
}

func main() {
	// Initialize database connection
	db, err := gorm.Open(sqlite.Open("kitab.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Migrate the schema & setup tables
	db.AutoMigrate(&Note{})
	db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS notesearch using fts5(noteid, title, content)")

	// Setup goldmark
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	// Setup server object
	server := Server{DB: db, MD: md}

	// Setup router
	router := gin.Default()
	router.SetTrustedProxies([]string{"localhost"})
	router.LoadHTMLGlob("templates/**/*.tmpl")
	router.Static("/assets", "./assets")

	router.GET("/", server.Index)
	router.GET("/notes/:id", server.GetNoteById)
	router.GET("/notes/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "notes/new.tmpl", gin.H{})
	})
	router.POST("/notes/new", server.NewNote)
	router.POST("/notes/search", server.SearchNote)

	// Listen and serve on 0.0.0.0:8080
	router.Run(":8080")
}

func (s *Server) Index(c *gin.Context) {
	var notes []Note
	result := s.DB.Select("id", "title").Find(&notes)

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
}

func (s *Server) GetNoteById(c *gin.Context) {
	var note Note
	id := c.Param("id")
	result := s.DB.Find(&note, id)

	if result.Error == nil {
		var buf bytes.Buffer
		if err := s.MD.Convert([]byte(note.Content), &buf); err != nil {
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
}

func (s *Server) NewNote(c *gin.Context) {
	var form NoteForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note := Note{Title: form.Title, Content: form.Content}
	result := s.DB.Create(&note)

	if result.RowsAffected > 0 && result.Error == nil {
		// Save to virtual table for easier searching of content
		s.DB.Exec(
			"INSERT INTO notesearch VALUES(?, ?, ?)",
			note.ID, note.Title, note.Content,
		)
		// Render markdown
		var buf bytes.Buffer
		if err := s.MD.Convert([]byte(note.Content), &buf); err != nil {
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
}

func (s *Server) SearchNote(c *gin.Context) {
	type SearchForm struct {
		Query string `form:"query" binding:"required"`
	}
	var searchForm SearchForm
	if err := c.ShouldBind(&searchForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var notes []Note
	searchResults := s.DB.Raw(`
		SELECT *
		FROM notesearch
		INNER JOIN notes ON notesearch.noteid = notes.id
		WHERE notesearch MATCH ?`,
		searchForm.Query,
	).Scan(&notes)

	if searchResults.Error == nil {
		c.HTML(http.StatusOK, "home/index.tmpl", gin.H{
			"count": len(notes),
			"notes": notes,
		})
		return
	}

	c.HTML(http.StatusInternalServerError, "shared/error.tmpl", gin.H{
		"msg":   "Something went wrong!",
		"error": searchResults.Error,
	})
}
