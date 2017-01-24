package model

import (
	"github.com/jinzhu/gorm"
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
