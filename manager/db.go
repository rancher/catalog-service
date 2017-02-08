package manager

import "github.com/rancher/catalog-service/model"

func (m *Manager) CreateConfigCatalogs() error {
	for name, config := range m.config {
		var catalogModel model.CatalogModel
		if err := m.db.FirstOrCreate(&catalogModel, &model.CatalogModel{
			Catalog: model.Catalog{
				Name:          name,
				URL:           config.URL,
				Branch:        config.Branch,
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

func (m *Manager) updateDb(catalog model.Catalog, templates []model.Template, versions []model.Version, newCommit string) error {
	tx := m.db.Begin()

	var catalogModel model.CatalogModel
	if err := tx.FirstOrCreate(&catalogModel, model.CatalogModel{
		Catalog: catalog,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&catalogModel).Update("commit", newCommit).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where(&model.TemplateModel{
		Template: model.Template{
			Catalog:       catalog.Name,
			EnvironmentId: catalog.EnvironmentId,
		},
	}).Delete(&model.TemplateModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where(&model.VersionModel{
		Version: model.Version{
			Catalog:       catalog.Name,
			EnvironmentId: catalog.EnvironmentId,
		},
	}).Delete(&model.VersionModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where(&model.FileModel{
		File: model.File{
			Catalog:       catalog.Name,
			EnvironmentId: catalog.EnvironmentId,
		},
	}).Delete(&model.FileModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, template := range templates {
		template.Catalog = catalog.Name
		template.EnvironmentId = catalog.EnvironmentId
		if err := tx.Create(&model.TemplateModel{
			Template: template,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, version := range versions {
		version.Catalog = catalog.Name
		version.EnvironmentId = catalog.EnvironmentId
		versionModel := model.VersionModel{
			Version: version,
		}
		if err := tx.Create(&versionModel).Error; err != nil {
			tx.Rollback()
			return err
		}
		for _, file := range version.Files {
			file.VersionID = versionModel.ID
			file.Catalog = version.Catalog
			file.EnvironmentId = version.EnvironmentId
			if err := tx.Create(&model.FileModel{
				File: file,
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}
