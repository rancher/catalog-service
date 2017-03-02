package manager

import (
	"github.com/jinzhu/gorm"
	"github.com/rancher/catalog-service/model"
)

func (m *Manager) removeCatalogsNotInConfig() error {
	var catalogs []model.CatalogModel
	m.db.Where(&model.CatalogModel{
		Catalog: model.Catalog{
			EnvironmentId: "global",
		},
	}).Find(&catalogs)
	for _, catalog := range catalogs {
		if _, ok := m.config[catalog.Name]; !ok {
			if err := m.db.Where(&model.CatalogModel{
				Catalog: model.Catalog{
					EnvironmentId: "global",
					Name:          catalog.Name,
				},
			}).Delete(&model.CatalogModel{}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) lookupEnvironmentCatalogs(environmentId string) ([]model.Catalog, error) {
	var catalogModels []model.CatalogModel
	if environmentId == "" {
		m.db.Find(&catalogModels)
	} else {
		m.db.Where(&model.CatalogModel{
			Catalog: model.Catalog{
				EnvironmentId: environmentId,
			},
		}).Find(&catalogModels)
	}
	var catalogs []model.Catalog
	for _, catalogModel := range catalogModels {
		if catalogModel.EnvironmentId != "global" {
			catalogs = append(catalogs, catalogModel.Catalog)
		}
	}
	return catalogs, nil
}

func (m *Manager) updateDb(tx *gorm.DB, catalog model.Catalog, templates []model.Template, newCommit string) error {
	if err := tx.Where(&model.CatalogModel{
		Catalog: model.Catalog{
			Name:          catalog.Name,
			EnvironmentId: catalog.EnvironmentId,
		},
	}).Delete(&model.CatalogModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	var catalogModel model.CatalogModel
	catalogModel.Catalog = catalog
	catalogModel.Commit = newCommit
	if err := tx.Create(&catalogModel).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, template := range templates {
		template.CatalogId = catalogModel.ID
		template.EnvironmentId = catalogModel.EnvironmentId
		templateModel := model.TemplateModel{
			Template: template,
		}
		if err := tx.Create(&templateModel).Error; err != nil {
			tx.Rollback()
			return err
		}

		for k, v := range template.Labels {
			if err := tx.Create(&model.TemplateLabelModel{
				TemplateLabel: model.TemplateLabel{
					TemplateId: templateModel.ID,
					Key:        k,
					Value:      v,
				},
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		if template.Category != "" {
			// TODO: TemplateCategory composite key
			template.Categories = append(template.Categories, template.Category)
		}

		for _, category := range template.Categories {
			var categoryModels []model.CategoryModel
			tx.Where(&model.CategoryModel{
				Category: model.Category{
					Name: category,
				},
			}).Find(&categoryModels)

			var categoryModel model.CategoryModel

			categoryFound := false
			for _, dbCategoryModel := range categoryModels {
				if dbCategoryModel.Name == category {
					categoryFound = true
					categoryModel = dbCategoryModel
					break
				}
			}

			if !categoryFound {
				categoryModel = model.CategoryModel{
					Category: model.Category{
						Name: category,
					},
				}
				if err := tx.Create(&categoryModel).Error; err != nil {
					tx.Rollback()
					return err
				}
			}

			if err := tx.Create(&model.TemplateCategoryModel{
				TemplateCategory: model.TemplateCategory{
					TemplateId: templateModel.ID,
					CategoryId: categoryModel.ID,
				},
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		for _, version := range template.Versions {
			version.TemplateId = templateModel.ID
			versionModel := model.VersionModel{
				Version: version,
			}
			if err := tx.Create(&versionModel).Error; err != nil {
				tx.Rollback()
				return err
			}

			for k, v := range version.Labels {
				if err := tx.Create(&model.VersionLabelModel{
					VersionLabel: model.VersionLabel{
						VersionId: versionModel.ID,
						Key:       k,
						Value:     v,
					},
				}).Error; err != nil {
					tx.Rollback()
					return err
				}
			}

			for _, file := range version.Files {
				file.VersionId = versionModel.ID
				if err := tx.Create(&model.FileModel{
					File: file,
				}).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	return tx.Commit().Error
}
