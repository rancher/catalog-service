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

func (m *Manager) lookupCatalogs(environmentId string) ([]model.Catalog, error) {
	var catalogModels []model.CatalogModel
	if environmentId == "" {
		if err := m.db.Find(&catalogModels).Error; err != nil {
			return nil, err
		}
	} else {
		if err := m.db.Where(&model.CatalogModel{
			Catalog: model.Catalog{
				EnvironmentId: environmentId,
			},
		}).Find(&catalogModels).Error; err != nil {
			return nil, err
		}
	}
	var catalogs []model.Catalog
	for _, catalogModel := range catalogModels {
		catalogs = append(catalogs, catalogModel.Catalog)
	}
	return catalogs, nil
}

func (m *Manager) lookupCatalog(environmentId, name string) (model.CatalogModel, error) {
	var catalogModel model.CatalogModel
	if err := m.db.Where(&model.CatalogModel{
		Catalog: model.Catalog{
			EnvironmentId: environmentId,
			Name:          name,
		},
	}).First(&catalogModel).Error; err != nil {
		return model.CatalogModel{}, err
	}
	return catalogModel, nil
}

func (m *Manager) deleteTemplates(templates []model.TemplateModel, templateCommits, templateLookup map[string]bool, tx *gorm.DB) error {
	for _, dbTemplate := range templates {
		// template either changed or is deleted
		if _, ok := templateCommits[dbTemplate.Commit]; !ok {
			if _, ok := templateLookup[dbTemplate.FolderName]; !ok {
				// delete from db only when template is deleted
				if err := tx.Where(&dbTemplate).Delete(&model.TemplateModel{}).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}
	return nil
}

func (m *Manager) lookupTemplates(catalogId uint, environmentId string) ([]model.TemplateModel, error) {
	var templates []model.TemplateModel
	if err := m.db.Where(&model.TemplateModel{
		Template: model.Template{
			CatalogId:     catalogId,
			EnvironmentId: environmentId,
		},
	}).Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (m *Manager) lookupVersions(templateIDs []uint) ([]model.VersionModel, error) {
	var versions []model.VersionModel
	versionQuery := `
	SELECT *
	FROM catalog_version
	WHERE template_id IN (?)`
	if err := m.db.Raw(versionQuery, templateIDs).Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}

func (m *Manager) saveCatalog(dbCatalogId uint, catalog model.Catalog, commit string, tx *gorm.DB) (model.CatalogModel, error) {
	catalogModel := model.CatalogModel{}
	catalogModel.Catalog = catalog
	catalogModel.Commit = commit

	if dbCatalogId == 0 {
		if err := tx.Create(&catalogModel).Error; err != nil {
			tx.Rollback()
			return model.CatalogModel{}, err
		}

		if err := tx.Where(&model.CatalogModel{
			Catalog: model.Catalog{
				Name:          catalog.Name,
				EnvironmentId: catalog.EnvironmentId,
			},
		}).Find(&catalogModel).Error; err != nil {
			return model.CatalogModel{}, err
		}

	} else {
		catalogModel.ID = dbCatalogId
		if err := tx.Model(&model.CatalogModel{}).Where(&model.CatalogModel{
			Base: model.Base{
				ID: catalogModel.ID,
			},
		}).Update(&catalogModel).Error; err != nil {
			tx.Rollback()
			return model.CatalogModel{}, err
		}
	}

	return catalogModel, nil
}

func (m *Manager) saveTemplates(templates []model.Template, environmentID string, catalogID uint, dbTemplateCommits map[string]bool, dbTemplateLookup map[string]model.TemplateModel, dbVersionCommits map[string]bool, tx *gorm.DB) error {
	for _, template := range templates {
		template.CatalogId = catalogID
		template.EnvironmentId = environmentID
		templateModel := model.TemplateModel{
			Template: template,
		}

		if _, ok := dbTemplateCommits[template.Commit]; !ok {
			if _, ok := dbTemplateLookup[template.FolderName]; ok {
				if templateModel.ID != 0 {
					if err := tx.Delete(&model.TemplateLabelModel{
						TemplateLabel: model.TemplateLabel{
							TemplateId: templateModel.ID,
						},
					}).Error; err != nil {
						tx.Rollback()
						return err
					}

					if err := tx.Delete(&model.TemplateCategoryModel{
						TemplateCategory: model.TemplateCategory{
							TemplateId: templateModel.ID,
						},
					}).Error; err != nil {
						tx.Rollback()
						return err
					}
				}

				// existing template, update
				templateModel.ID = dbTemplateLookup[template.FolderName].ID
				if err := tx.Model(&model.TemplateModel{}).Where(&model.TemplateModel{
					Template: model.Template{
						CatalogId:  catalogID,
						FolderName: template.FolderName,
					},
				}).Update(&templateModel).Error; err != nil {
					tx.Rollback()
					return err
				}

			} else {
				// new template, add to db
				if err := tx.Create(&templateModel).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
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
			if _, ok := dbVersionCommits[version.Commit]; !ok {
				// add version to db
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
	}
	return nil
}

func (m *Manager) deleteVersions(versions []model.VersionModel, versionCommits map[string]bool, tx *gorm.DB) error {
	for _, dbVersion := range versions {
		if _, ok := versionCommits[dbVersion.Commit]; !ok {
			// delete dbVersion bc version changed
			if err := tx.Where(&dbVersion).Delete(&model.VersionModel{}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return nil
}

func (m *Manager) updateDb(catalog model.Catalog, templates []model.Template, newCommit string) error {
	tx := m.db.Begin()

	// Pick out db Catalog ID
	dbCatalog, err := m.lookupCatalog(catalog.EnvironmentId, catalog.Name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			dbCatalog, err = m.saveCatalog(0, catalog, newCommit, tx)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Pick out db Template IDs
	dbTemplates, err := m.lookupTemplates(dbCatalog.ID, catalog.EnvironmentId)
	if err != nil {
		return err
	}

	dbTemplateIds := []uint{}
	dbTemplateCommits := make(map[string]bool)
	dbTemplateLookup := make(map[string]model.TemplateModel)
	for _, dbTemplate := range dbTemplates {
		dbTemplateIds = append(dbTemplateIds, dbTemplate.ID)
		dbTemplateCommits[dbTemplate.Commit] = true
		dbTemplateLookup[dbTemplate.FolderName] = dbTemplate
	}

	// Pick out db Version IDs
	dbVersions, err := m.lookupVersions(dbTemplateIds)
	if err != nil {
		return err
	}

	dbVersionCommits := make(map[string]bool)
	for _, dbVersion := range dbVersions {
		dbVersionCommits[dbVersion.Commit] = true
	}

	templateCommits := make(map[string]bool)
	templateLookup := make(map[string]bool)
	versionCommits := make(map[string]bool)

	for _, template := range templates {
		templateCommits[template.Commit] = true
		templateLookup[template.FolderName] = true
		for _, version := range template.Versions {
			versionCommits[version.Commit] = true
		}
	}

	catalogModel, err := m.saveCatalog(dbCatalog.ID, catalog, newCommit, tx)
	if err != nil {
		return err
	}

	if err := m.deleteTemplates(dbTemplates, templateCommits, templateLookup, tx); err != nil {
		return err
	}

	if err := m.deleteVersions(dbVersions, versionCommits, tx); err != nil {
		return err
	}

	if err := m.saveTemplates(templates, catalogModel.EnvironmentId, catalogModel.ID, dbTemplateCommits, dbTemplateLookup, dbVersionCommits, tx); err != nil {
		return err
	}

	return tx.Commit().Error
}
