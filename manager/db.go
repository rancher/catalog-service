package manager

import "github.com/rancher/catalog-service/model"

func (m *Manager) CreateConfigCatalogs() error {
	for name, config := range m.config {
		var catalogModel model.CatalogModel
		if err := m.db.FirstOrCreate(&catalogModel, &model.CatalogModel{
			Catalog: model.Catalog{
				Name:          name,
				URL:           config.URL,
				EnvironmentId: "global",
			},
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) lookupCatalogs(environmentId string) ([]model.Catalog, error) {
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
		catalogs = append(catalogs, catalogModel.Catalog)
	}
	return catalogs, nil
}

func (m *Manager) updateDb(catalog model.Catalog, templates []model.Template, newCommit string) error {
	tx := m.db.Begin()

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
			if err := tx.Create(&model.LabelModel{
				Label: model.Label{
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
			var categoryModel model.CategoryModel
			if err := tx.FirstOrCreate(&categoryModel, model.CategoryModel{
				Category: model.Category{
					Name: category,
				},
			}).Error; err != nil {
				tx.Rollback()
				return err
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
			version.EnvironmentId = catalog.EnvironmentId
			versionModel := model.VersionModel{
				Version: version,
			}
			if err := tx.Create(&versionModel).Error; err != nil {
				tx.Rollback()
				return err
			}

			for _, file := range version.Files {
				file.VersionId = versionModel.ID
				file.EnvironmentId = version.EnvironmentId
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
