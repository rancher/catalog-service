package manager

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/rancher/catalog-service/model"
)

type Manager struct {
	cacheRoot  string
	configFile string
	config     map[string]CatalogConfig
	db         *gorm.DB
}

func NewManager(cacheRoot string, configFile string, db *gorm.DB) *Manager {
	return &Manager{
		cacheRoot:  cacheRoot,
		configFile: configFile,
		db:         db,
	}
}

func (m *Manager) RefreshAll() error {
	if err := m.readConfig(); err != nil {
		return err
	}
	if err := m.CreateConfigCatalogs(); err != nil {
		return err
	}
	catalogs, err := m.lookupCatalogs("")
	if err != nil {
		return err
	}
	for _, catalog := range catalogs {
		if err := m.refreshCatalog(catalog); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Refresh(environmentId string) error {
	if environmentId == "global" {
		if err := m.readConfig(); err != nil {
			return err
		}
		if err := m.CreateConfigCatalogs(); err != nil {
			return err
		}
	}
	catalogs, err := m.lookupCatalogs(environmentId)
	if err != nil {
		return err
	}
	for _, catalog := range catalogs {
		if err := m.refreshCatalog(catalog); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) refreshCatalog(catalog model.Catalog) error {

	repoPath, commit, err := m.prepareRepoPath(catalog)
	if err != nil {
		return err
	}

	// Catalog is already up to date
	if commit == catalog.Commit {
		log.Debugf("Catalog %s is already up to date", catalog.Name)
		return nil
	}

	templates, err := traverseFiles(repoPath)
	if err != nil {
		return err
	}

	log.Debugf("Updating catalog %s", catalog.Name)

	return m.updateDb(catalog, templates, commit)
}
