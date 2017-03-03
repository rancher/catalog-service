package manager

import (
	"database/sql"

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

func (m *Manager) lookupCatalog(environmentId, name string) (model.Catalog, error) {
	var catalogModel model.CatalogModel
	if err := m.db.Where(&model.CatalogModel{
		Catalog: model.Catalog{
			EnvironmentId: environmentId,
			Name:          name,
		},
	}).First(&catalogModel).Error; err != nil {
		return model.Catalog{}, err
	}
	return catalogModel.Catalog, nil
}

func (m *Manager) updateDb(catalog model.Catalog, templates []model.Template, newCommit string) error {
	tx, err := m.db.DB().Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(`
DELETE FROM catalog
WHERE catalog.name = ?
AND catalog.environment_id = ?
`, catalog.Name, catalog.EnvironmentId); err != nil {
		tx.Rollback()
		return err
	}

	catalogResult, err := tx.Exec(`
INSERT INTO catalog (environment_id, name, url, branch, [commit])
VALUES (?, ?, ?, ?, ?)
`, catalog.EnvironmentId, catalog.Name, catalog.URL, catalog.Branch, newCommit)
	if err != nil {
		tx.Rollback()
		return err
	}

	catalogId, err := catalogResult.LastInsertId()
	if err != nil {
		return err
	}

	for _, template := range templates {
		templateResult, err := tx.Exec(`
INSERT INTO catalog_template (environment_id, catalog_id, name, is_system, description, default_version, path, maintainer, license, project_url, upgrade_from, folder_name, base, icon, icon_filename, readme)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, catalog.EnvironmentId, catalogId, template.Name, template.IsSystem, template.Description, template.DefaultVersion, template.Path, template.Maintainer, template.License, template.ProjectURL, template.UpgradeFrom, template.FolderName, template.Base, template.Icon, template.IconFilename, template.Readme)
		if err != nil {
			tx.Rollback()
			return err
		}

		templateId, err := templateResult.LastInsertId()
		if err != nil {
			return err
		}

		for k, v := range template.Labels {
			if _, err := tx.Exec(`
INSERT INTO catalog_label (template_id, key, value)
VALUES (?, ?, ?)
`, templateId, k, v); err != nil {
				tx.Rollback()
				return err
			}
		}

		if template.Category != "" {
			// TODO: TemplateCategory composite key
			template.Categories = append(template.Categories, template.Category)
		}

		for _, category := range template.Categories {
			categoryId, err := m.ensureCategoryExists(tx, category)
			if err != nil {
				tx.Rollback()
				return err
			}

			if _, err := tx.Exec(`
INSERT INTO catalog_template_category (template_id, category_id)
VALUES (?, ?)
`, templateId, categoryId); err != nil {
				tx.Rollback()
				return err
			}
		}

		for _, version := range template.Versions {
			versionResult, err := tx.Exec(`
INSERT INTO catalog_version (template_id, revision, version, minimum_rancher_version, maximum_rancher_version, upgrade_from, readme)
VALUES (?, ?, ?, ?, ?, ?, ?)
`, templateId, version.Revision, version.Version, version.MinimumRancherVersion, version.MaximumRancherVersion, version.UpgradeFrom, version.Readme)
			if err != nil {
				tx.Rollback()
				return err
			}

			versionId, err := versionResult.LastInsertId()
			if err != nil {
				return err
			}

			for k, v := range version.Labels {
				if _, err := tx.Exec(`
INSERT INTO catalog_version_label (version_id, key, value)
VALUES (?, ?, ?)
`, versionId, k, v); err != nil {
					tx.Rollback()
					return err
				}
			}

			for _, file := range version.Files {
				if _, err := tx.Exec(`
INSERT INTO catalog_file (version_id, name, contents)
VALUES (?, ?, ?)
`, versionId, file.Name, file.Contents); err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	return tx.Commit()
}

func (m *Manager) ensureCategoryExists(tx *sql.Tx, category string) (int64, error) {
	rows, err := tx.Query(`
SELECT id, name
FROM catalog_category
WHERE catalog_category.name = ?
`, category)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var foundCategory string
		var foundCategoryId int64
		if err = rows.Scan(&foundCategoryId, &foundCategory); err != nil {
			return 0, err
		}
		if foundCategory == category {
			return foundCategoryId, nil
		}
	}
	if err = rows.Err(); err != nil {
		return 0, err
	}

	categoryResult, err := tx.Exec(`
INSERT INTO catalog_category (name)
VALUES (?)
`, category)
	if err != nil {
		return 0, err
	}

	return categoryResult.LastInsertId()
}
