package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type File struct {
	Name      string `json:"name"`
	Contents  string
	VersionID uint
}

type FileModel struct {
	gorm.Model
	File
}
