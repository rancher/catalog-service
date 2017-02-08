package model

import (
	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/client"
)

type Catalog struct {
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	Branch               string `json:"branch"`
	Commit               string `json:"commit"`
	EnvironmentId        string `json:"environmentId"`
	TransitioningMessage string `json:"transitioningMessage"`
}

type CatalogModel struct {
	Base
	Catalog
}

type CatalogResource struct {
	client.Resource
	Catalog
}

type CatalogCollection struct {
	client.Collection
	Data []CatalogResource `json:"data,omitempty"`
}

func LookupCatalog(db *gorm.DB, environmentId, name string) *Catalog {
	var catalogModel CatalogModel
	db.Where(&CatalogModel{
		Catalog: Catalog{
			Name: name,
		},
	}).Where("environment_id = ? OR environment_id = ?", environmentId, "global").First(&catalogModel)
	return &catalogModel.Catalog
}

func LookupCatalogs(db *gorm.DB, environmentId string) []Catalog {
	var catalogModels []CatalogModel
	db.Where("environment_id = ? OR environment_id = ?", environmentId, "global").Find(&catalogModels)
	var catalogs []Catalog
	for _, catalogModel := range catalogModels {
		catalogs = append(catalogs, catalogModel.Catalog)
	}
	return catalogs
}

// TODO: return error
func DeleteCatalog(db *gorm.DB, environmentId, name string) {
	tx := db.Begin()

	if err := tx.Where(&CatalogModel{
		Catalog: Catalog{
			Name:          name,
			EnvironmentId: environmentId,
		},
	}).Delete(&CatalogModel{}).Error; err != nil {
		tx.Rollback()
	}

	if err := tx.Where(&TemplateModel{
		Template: Template{
			Catalog:       name,
			EnvironmentId: environmentId,
		},
	}).Delete(&TemplateModel{}).Error; err != nil {
		tx.Rollback()
	}

	if err := tx.Where(&VersionModel{
		Version: Version{
			Catalog:       name,
			EnvironmentId: environmentId,
		},
	}).Delete(&VersionModel{}).Error; err != nil {
		tx.Rollback()
	}

	tx.Commit()
}
