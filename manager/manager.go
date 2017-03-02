package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/rancher/catalog-service/model"
)

type Manager struct {
	cacheRoot  string
	configFile string
	config     map[string]CatalogConfig
	strict     bool
	db         *gorm.DB
}

func NewManager(cacheRoot string, configFile string, strict bool, db *gorm.DB) *Manager {
	return &Manager{
		cacheRoot:  cacheRoot,
		configFile: configFile,
		strict:     strict,
		db:         db,
	}
}

func (m *Manager) RefreshAll() error {
	if err := m.refreshConfigCatalogs(); err != nil {
		return err
	}
	catalogs, err := m.lookupEnvironmentCatalogs("")
	if err != nil {
		return err
	}
	for _, catalog := range catalogs {
		if err := m.refreshCatalog(catalog, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Refresh(environmentId string) error {
	if environmentId == "global" {
		return m.refreshConfigCatalogs()
	}
	catalogs, err := m.lookupEnvironmentCatalogs(environmentId)
	if err != nil {
		return err
	}
	for _, catalog := range catalogs {
		if err := m.refreshCatalog(catalog, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) refreshConfigCatalogs() error {
	if err := m.readConfig(); err != nil {
		return err
	}
	if err := m.removeCatalogsNotInConfig(); err != nil {
		return err
	}

	for name, config := range m.config {
		if err := m.refreshCatalog(model.Catalog{
			Name:          name,
			URL:           config.URL,
			Branch:        config.Branch,
			EnvironmentId: "global",
		}, func() (*gorm.DB, error) {
			tx := m.db.Begin()

			var catalogModel model.CatalogModel
			if err := tx.FirstOrCreate(&catalogModel, &model.CatalogModel{
				Catalog: model.Catalog{
					Name:          name,
					URL:           config.URL,
					Branch:        config.Branch,
					EnvironmentId: "global",
				},
			}).Error; err != nil {
				return nil, err
			}

			return tx, nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) refreshCatalog(catalog model.Catalog, txBegin func() (*gorm.DB, error)) error {
	repoPath, commit, err := m.prepareRepoPath(catalog)
	if err != nil {
		return err
	}

	// Catalog is already up to date
	if commit == catalog.Commit {
		log.Debugf("Catalog %s is already up to date", catalog.Name)
		return nil
	}

	templates, errors, err := traverseFiles(repoPath)
	if err != nil {
		return err
	}
	if errors != nil {
		if m.strict {
			return fmt.Errorf("%v", errors)
		}
		log.Errorf("Errors while parsing repo: %v", errors)
	}

	log.Debugf("Updating catalog %s", catalog.Name)

	var tx *gorm.DB
	if txBegin == nil {
		tx = m.db.Begin()
	} else {
		tx, err = txBegin()
		if err != nil {
			return err
		}
	}

	return m.updateDb(tx, catalog, templates, commit)
}
