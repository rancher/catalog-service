package model

import (
	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/client"
)

// TODO: might need a Base field for filtering
// TODO: might need a FolderName field for filtering
type Version struct {
	Catalog               string `json:"catalogId"`
	EnvironmentId         string `json:"environmentId"`
	Template              string `json:"template"`
	Revision              int    `json:"revision"`
	Version               string `json:"version"`
	MinimumRancherVersion string `json:"minimumRancherVersion" yaml:"minimum_rancher_version"`
	MaximumRancherVersion string `json:"maximumRancherVersion" yaml:"maximum_rancher_version"`
	UpgradeFrom           string `json:"upgradeFrom" yaml:"upgrade_from"`

	// TODO move to model
	Files     []File
	Questions []Question
	Readme    string `json:"readme"`
}

type VersionModel struct {
	Base
	Version
}

type TemplateVersionResource struct {
	client.Resource
	Version

	Bindings            map[string]Bindings `json:"bindings"`
	Files               map[string]string   `json:"files"`
	Questions           []Question          `json:"questions"`
	UpgradeVersionLinks map[string]string   `json:"upgradeVersionLinks"`
}

// TODO: needs a base filter (make sure to use a map)
func LookupVersionModel(db *gorm.DB, environmentId, catalog, template string, revision int) *VersionModel {
	var versionModel VersionModel
	db.Where(&VersionModel{
		Version: Version{
			Catalog:  catalog,
			Template: template,
		},
	}).Where(map[string]interface{}{
		"revision": revision,
	}).Where("environment_id = ? OR environment_id = ?", environmentId, "global").First(&versionModel)
	return &versionModel
}

func LookupVersions(db *gorm.DB, environmentId, catalog, template string) []Version {
	var versionModels []VersionModel
	db.Where(&VersionModel{
		Version: Version{
			Catalog:  catalog,
			Template: template,
		},
	}).Where("environment_id = ? OR environment_id = ?", environmentId, "global").Find(&versionModels)
	var versions []Version
	for _, versionModel := range versionModels {
		versions = append(versions, versionModel.Version)
	}
	return versions
}
