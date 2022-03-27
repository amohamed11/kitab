package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Note struct {
	gorm.Model
	Title   string
	Content string
}

func Init() {
	var err error
	db, err = gorm.Open(sqlite.Open("kitab.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// Setup DB schema
	db.AutoMigrate(&Note{})
	db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS notesearch using fts5(noteid, title, content, tokenize="porter");`)
	db.Exec(`
	  CREATE TRIGGER  add_notesearch
	     AFTER INSERT ON notes
	  BEGIN
	     INSERT INTO notesearch(noteid, title, content) VALUES(NEW.id, NEW.title, NEW.content);
	  END;
	`)
	db.Exec(`
	  CREATE TRIGGER  update_notesearch
	     AFTER UPDATE ON notes
	     WHEN OLD.title <> NEW.title OR OLD.content <> NEW.content
	  BEGIN
	     UPDATE notesearch
	     SET title=NEW.title, content=NEW.content
	     WHERE noteid = OLD.id;
	  END;
	`)
}

func GetDB() *gorm.DB {
	return db
}
