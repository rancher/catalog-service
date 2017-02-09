package model

import "github.com/jinzhu/gorm"

type File struct {
	EnvironmentId string `json:"environmentId"`
	VersionId     uint   `sql:"type:integer REFERENCES catalog_version(id) ON DELETE CASCADE"`

	Name     string `json:"name"`
	Contents string
}

type FileModel struct {
	Base
	File
}

func LookupFiles(db *gorm.DB, catalog, environmentId string, versionId uint) []File {
	var fileModels []FileModel
	db.Raw(`
SELECT catalog_file.*
FROM catalog_file
WHERE (catalog_file.environment_id = ? OR catalog_file.environment_id = ?)
AND catalog_file.version_id = ?
`, environmentId, "global", versionId).Scan(&fileModels)

	var files []File
	for _, fileModel := range fileModels {
		files = append(files, fileModel.File)
	}
	return files
}
