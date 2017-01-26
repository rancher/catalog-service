package model

import (
	"github.com/jinzhu/gorm"
)

type File struct {
	Catalog       string `json:"catalogId"`
	EnvironmentId string `json:"environmentId"`
	Name          string `json:"name"`
	Contents      string
	VersionID     uint
}

type FileModel struct {
	gorm.Model
	File
}
