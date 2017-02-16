package model

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/client"
)

type Template struct {
	EnvironmentId string `json:"environmentId"`
	CatalogId     uint   `sql:"type:integer REFERENCES catalog(id) ON DELETE CASCADE"`

	Name           string `json:"name"`
	IsSystem       string `json:"isSystem"`
	Description    string `json:"description"`
	DefaultVersion string `json:"defaultVersion" yaml:"version"`
	Path           string `json:"path"`
	Maintainer     string `json:"maintainer"`
	License        string `json:"license"`
	ProjectURL     string `json:"projectURL" yaml:"projectURL"`
	UpgradeFrom    string `json:"upgradeFrom"`
	FolderName     string `json:"folderName"`
	Catalog        string `json:"catalogId"`
	Base           string `json:"templateBase"`
	Icon           string `json:"icon"`
	IconFilename   string `json:"iconFilename"`
	Readme         string `json:"readme"`

	Categories []string          `sql:"-" json:"categories"`
	Labels     map[string]string `sql:"-" json:"labels"`

	Versions []Version `sql:"-"`
	Category string    `sql:"-"`
}

type TemplateModel struct {
	Base
	Template
}

type TemplateResource struct {
	client.Resource
	Template

	VersionLinks map[string]string `json:"versionLinks"`
}

type TemplateCollection struct {
	client.Collection
	Data []TemplateResource `json:"data,omitempty"`
}

func LookupTemplate(db *gorm.DB, environmentId, catalog, folderName, base string) *Template {
	var templateModel TemplateModel
	db.Raw(`
SELECT catalog_template.*
FROM catalog_template, catalog
WHERE (catalog_template.environment_id = ? OR catalog_template.environment_id = ?)
AND catalog_template.catalog_id = catalog.id
AND catalog.name = ?
AND catalog_template.base = ?
AND catalog_template.folder_name = ?
`, environmentId, "global", catalog, base, folderName).Scan(&templateModel)

	fillInTemplate(db, &templateModel)
	return &templateModel.Template
}

func fillInTemplate(db *gorm.DB, templateModel *TemplateModel) {
	catalog := GetCatalog(db, templateModel.CatalogId)
	if catalog != nil {
		templateModel.Catalog = catalog.Name
	}
	templateModel.Categories = lookupTemplateCategories(db, templateModel.ID)
	templateModel.Labels = lookupTemplateLabels(db, templateModel.ID)
	templateModel.Versions = lookupVersions(db, templateModel.ID)
}

func LookupTemplates(db *gorm.DB, environmentId, catalog, templateBaseEq string, categories, categoriesNe []string) []Template {
	var templateModels []TemplateModel

	params := []interface{}{environmentId, "global"}
	if catalog != "" {
		params = append(params, catalog)
	}
	if templateBaseEq != "" {
		params = append(params, templateBaseEq)
	}
	for _, category := range categories {
		params = append(params, category)
	}
	for _, categoryNe := range categoriesNe {
		params = append(params, categoryNe)
	}

	var query string
	if len(categories) == 0 && len(categoriesNe) == 0 {
		query = `
SELECT catalog_template.*
FROM catalog_template, catalog
WHERE (catalog_template.environment_id = ? OR catalog_template.environment_id = ?)
AND catalog_template.catalog_id = catalog.id`
	} else {
		query = `
SELECT catalog_template.*
FROM catalog_template, catalog_template_category, catalog_category, catalog
WHERE (catalog_template.environment_id = ? OR catalog_template.environment_id = ?)
AND catalog_template.id = catalog_template_category.template_id
AND catalog_category.id = catalog_template_category.category_id
AND catalog_template.catalog_id = catalog.id`
	}

	if catalog != "" {
		query += `
AND catalog.name = ?`
	}
	if templateBaseEq != "" {
		query += `
AND catalog_template.base = ?`
	}
	if len(categories) > 0 {
		query += fmt.Sprintf(`
AND catalog_category.name IN (%s)`, listQuery(len(categories)))
	}
	if len(categoriesNe) > 0 {
		query += fmt.Sprintf(`
AND catalog_category.name NOT IN (%s)`, listQuery(len(categoriesNe)))
	}

	db.Raw(query, params...).Find(&templateModels)

	var templates []Template
	for _, templateModel := range templateModels {
		fillInTemplate(db, &templateModel)
		templates = append(templates, templateModel.Template)
	}
	return templates
}

func listQuery(size int) string {
	var query string
	for i := 0; i < size; i++ {
		query += " ? ,"
	}
	return fmt.Sprintf("(%s)", strings.TrimSuffix(query, ","))
}
