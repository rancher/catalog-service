package model

import (
	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/client"
)

const (
	baseVersionQuery = `SELECT catalog_version.*
FROM catalog_version, catalog_template, catalog
WHERE (catalog.environment_id = ? OR catalog.environment_id = ?)
AND catalog_version.template_id = catalog_template.id
AND catalog_template.catalog_id = catalog.id
AND catalog.name = ?
AND catalog_template.base = ?
AND catalog_template.folder_name = ?`
)

type Version struct {
	TemplateId uint `sql:"type:integer REFERENCES catalog_template(id) ON DELETE CASCADE"`

	Revision                       *int   `json:"revision"`
	Version                        string `json:"version"`
	MinimumRancherVersion          string
	MinimumRancherVersionCamelCase string `json:"minimumRancherVersion" yaml:"minimumRancherVersion"`
	MinimumRancherVersionSnakeCase string `json:"minimumRancherVersion" yaml:"minimum_rancher_version"`
	MaximumRancherVersion          string
	MaximumRancherVersionCamelCase string `json:"maximumRancherVersion" yaml:"maximumRancherVersion"`
	MaximumRancherVersionSnakeCase string `json:"maximumRancherVersion" yaml:"maximum_rancher_version"`
	UpgradeFrom                    string
	UpgradeFromCamelCase           string `json:"upgradeFrom" yaml:"upgradeFrom"`
	UpgradeFromSnakeCase           string `json:"upgradeFrom" yaml:"upgrade_from"`
	Readme                         string `json:"readme"`

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
	TemplateId          string              `json:"templateId"`
}

func LookupVersionByRevision(db *gorm.DB, environmentId, catalog, base, template string, revision int) *Version {
	var versionModel VersionModel
	if err := db.Raw(baseVersionQuery+`
AND catalog_version.revision = ?
`, environmentId, "global", catalog, base, template, revision).Scan(&versionModel).Error; err == gorm.ErrRecordNotFound {
		return nil
	}

	versionModel.Labels = lookupVersionLabels(db, versionModel.ID)
	versionModel.Files = lookupFiles(db, versionModel.ID)

	return &versionModel.Version
}

func LookupVersionByVersion(db *gorm.DB, environmentId, catalog, base, template string, version string) *Version {
	var versionModel VersionModel
	if err := db.Raw(baseVersionQuery+`
AND catalog_version.version = ?
`, environmentId, "global", catalog, base, template, version).Scan(&versionModel).Error; err == gorm.ErrRecordNotFound {
		return nil
	}

	versionModel.Labels = lookupVersionLabels(db, versionModel.ID)
	versionModel.Files = lookupFiles(db, versionModel.ID)

	return &versionModel.Version
}

func lookupVersions(db *gorm.DB, templateId uint) []Version {
	var versionModels []VersionModel

	query := `
	SELECT *
	FROM catalog_version
	WHERE template_id = ?`

	db.Raw(query, templateId).Find(&versionModels)

	var versions []Version
	for _, versionModel := range versionModels {
		versionModel.Labels = lookupVersionLabels(db, versionModel.ID)
		versionModel.Files = lookupFiles(db, versionModel.ID)
		versions = append(versions, versionModel.Version)
	}
	return versions
}
