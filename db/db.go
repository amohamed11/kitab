package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

type Folder struct {
	gorm.Model
	Name  string
	Notes []Note
}

type Note struct {
	gorm.Model
	Title    string
	Content  string
	FolderID uint
}

func Init(db_url string) {
	var err error
	db, err = gorm.Open(sqlite.Open(db_url), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
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
