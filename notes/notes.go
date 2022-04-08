package notes

import (
	"bytes"
	"html/template"
	"net/http"

	"git.sr.ht/~anecdotal/kitab/db"
	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

type NoteForm struct {
	Title   string `form:"title"`
	Content string `form:"content"`
}

type NoteController struct{}

var MD goldmark.Markdown

func (n NoteController) Init() {
	// Setup goldmark
	MD = goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
}

func (n NoteController) Index(c *gin.Context) {
	var notes []db.Note
	result := db.GetDB().Select("id", "title").Find(&notes)

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

func (n NoteController) GetById(c *gin.Context) {
	var note db.Note
	id := c.Param("id")
	result := db.GetDB().Find(&note, id)

	if result.Error == nil {
		var buf bytes.Buffer
		if err := MD.Convert([]byte(note.Content), &buf); err != nil {
			panic(err)
		}
		c.HTML(http.StatusOK, "notes/edit.tmpl", gin.H{
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

func (n NoteController) New(c *gin.Context) {
	var form NoteForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note := db.Note{Title: form.Title, Content: form.Content}
	result := db.GetDB().Create(&note)

	if result.RowsAffected > 0 && result.Error == nil {
		c.HTML(http.StatusCreated, "notes/edit.tmpl", gin.H{
			"note": note,
		})
		return
	}

	c.HTML(http.StatusInternalServerError, "shared/error.tmpl", gin.H{
		"msg":   "Something went wrong!",
		"error": result.Error,
	})
}

func (n NoteController) Edit(c *gin.Context) {
	var form NoteForm
	id := c.Param("id")
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedNote db.Note
	result := db.GetDB().Model(&updatedNote).Where("id = ?", id).Updates(db.Note{Title: form.Title, Content: form.Content})

	if result.RowsAffected > 0 && result.Error == nil {
		c.Writer.Header().Set("HX-Trigger", `{"showToast":"success"}`)
		c.HTML(http.StatusOK, "notes/_editor.tmpl", gin.H{
			"note": updatedNote,
		})
		return
	}

	c.Writer.Header().Set("HX-Trigger", `{"showToast":"failure"}`)
	c.HTML(http.StatusInternalServerError, "notes/_editor.tmpl", gin.H{
		"error": result.Error,
	})
}

func (n NoteController) Search(c *gin.Context) {
	type SearchForm struct {
		Query string `form:"query" binding:"required"`
	}
	var searchForm SearchForm
	if err := c.ShouldBind(&searchForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var notes []db.Note
	searchResults := db.GetDB().Raw(`
		SELECT *
		FROM notesearch
		INNER JOIN notes ON notesearch.noteid = notes.id
		WHERE notesearch MATCH ?`,
		searchForm.Query,
	).Scan(&notes)

	if searchResults.Error == nil {
		c.HTML(http.StatusOK, "notes/list.tmpl", gin.H{
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
