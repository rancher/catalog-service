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

func LookupFiles(db *gorm.DB, catalog, environmentId string, versionId uint) []File {
	var fileModels []FileModel
	db.Where(&FileModel{
		File: File{
			Catalog:       catalog,
			EnvironmentId: environmentId,
			VersionID:     versionId,
		},
	}).Find(&fileModels)
	var files []File
	for _, fileModel := range fileModels {
		files = append(files, fileModel.File)
	}
	return files
}
