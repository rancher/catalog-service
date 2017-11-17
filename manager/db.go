package manager

import (
	"fmt"
	catalogClient "github.com/rancher/catalog-service/client"
	"github.com/rancher/catalog-service/model"
	catalogv1 "github.com/rancher/type/apis/catalog.cattle.io/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *Manager) removeCatalogsNotInConfig() error {
	catalogClient := catalogClient.CatalogClient
	catalogList, err := catalogClient.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, catalog := range catalogList.Items {
		if _, ok := m.config[catalog.Name]; !ok {
			if err := catalogClient.Delete(catalog.Name, &metav1.DeleteOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) lookupCatalogs(environmentId string) ([]catalogv1.Catalog, error) {
	catalogList, err := catalogClient.CatalogClient.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", model.ProjectLabel, environmentId),
	})
	if err != nil {
		return nil, err
	}
	catalogs := []catalogv1.Catalog{}
	for _, catalog := range catalogList.Items {
		catalogs = append(catalogs, catalog)
	}
	return catalogs, nil
}

func (m *Manager) lookupCatalog(environmentId, name string) (catalogv1.Catalog, error) {
	catalog, err := catalogClient.CatalogClient.Get(name, metav1.GetOptions{})
	if err != nil {
		return catalogv1.Catalog{}, err
	}
	return *catalog, nil
}

func (m *Manager) updateDb(catalog catalogv1.Catalog, templates []model.Template, newCommit string) error {
	tx := m.db.Begin()

	catalog.Spec.Commit = newCommit
	_, err := catalogClient.CatalogClient.Update(&catalog)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if errors.IsNotFound(err) {
		if _, err := catalogClient.CatalogClient.Create(&catalog); err != nil {
			return err
		}
	}

	query := `
	DELETE FROM catalog_template
	WHERE catalog_template.catalog_name = ?
	`
	if err := m.db.Exec(query, catalog.Name).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, template := range templates {
		template.CatalogName = catalog.Name
		template.EnvironmentId = catalog.Labels[model.ProjectLabel]
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
