package model

import (
	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/client"
)

type Version struct {
	TemplateId uint `sql:"type:integer REFERENCES catalog_template(id) ON DELETE CASCADE"`

	Revision              int    `json:"revision"`
	Version               string `json:"version"`
	MinimumRancherVersion string `json:"minimumRancherVersion" yaml:"minimum_rancher_version"`
	MaximumRancherVersion string `json:"maximumRancherVersion" yaml:"maximum_rancher_version"`
	UpgradeFrom           string `json:"upgradeFrom" yaml:"upgrade_from"`
	Readme                string `json:"readme"`

	Labels map[string]string `sql:"-" json:"labels"`

	Files     []File     `sql:"-"`
	Questions []Question `sql:"-"`
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

func LookupVersion(db *gorm.DB, environmentId, catalog, base, template string, revision int) *Version {
	var versionModel VersionModel
	db.Raw(`
SELECT catalog_version.*
FROM catalog_version, catalog_template, catalog
WHERE (catalog.environment_id = ? OR catalog.environment_id = ?)
AND catalog_version.template_id = catalog_template.id
AND catalog_template.catalog_id = catalog.id
AND catalog.name = ?
AND catalog_template.base = ?
AND catalog_template.folder_name = ?
AND catalog_version.revision = ?
`, environmentId, "global", catalog, base, template, revision).Scan(&versionModel)

	versionModel.Labels = lookupVersionLabels(db, versionModel.ID)
	versionModel.Files = lookupFiles(db, versionModel.ID)

	return &versionModel.Version
}

func lookupVersions(db *gorm.DB, templateId uint) []Version {
	var versionModels []VersionModel
	db.Where(&VersionModel{
		Version: Version{
			TemplateId: templateId,
		},
	}).Find(&versionModels)

	var versions []Version
	for _, versionModel := range versionModels {
		versionModel.Labels = lookupVersionLabels(db, versionModel.ID)
		versionModel.Files = lookupFiles(db, versionModel.ID)
		versions = append(versions, versionModel.Version)
	}
	return versions
}
