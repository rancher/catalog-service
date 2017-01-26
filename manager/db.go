package manager

import "github.com/rancher/catalog-service/model"

func (m *Manager) updateDb(name string, config CatalogConfig, templates []model.Template, versions []model.Version) error {
	tx := m.db.Begin()

	catalogQuery := model.CatalogModel{
		Catalog: model.Catalog{
			Name:          name,
			URL:           config.URL,
			Branch:        config.Branch,
			EnvironmentId: config.EnvironmentId,
		},
	}

	if err := tx.Where(&catalogQuery).Delete(&model.CatalogModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Create(&catalogQuery).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where(&model.TemplateModel{
		Template: model.Template{
			Catalog:       name,
			EnvironmentId: config.EnvironmentId,
		},
	}).Delete(&model.TemplateModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where(&model.VersionModel{
		Version: model.Version{
			Catalog:       name,
			EnvironmentId: config.EnvironmentId,
		},
	}).Delete(&model.VersionModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where(&model.FileModel{
		File: model.File{
			Catalog:       name,
			EnvironmentId: config.EnvironmentId,
		},
	}).Delete(&model.FileModel{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, template := range templates {
		template.Catalog = name
		template.EnvironmentId = config.EnvironmentId
		if err := tx.Create(&model.TemplateModel{
			Template: template,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, version := range versions {
		version.Catalog = name
		version.EnvironmentId = config.EnvironmentId
		versionModel := model.VersionModel{
			Version: version,
		}
		if err := tx.Create(&model.VersionModel{
			Version: version,
		}).Error; err != nil {
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
