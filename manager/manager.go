package manager

import (
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/rancher/catalog-service/model"
)

// TODO: move elsewhere
type CatalogConfig struct {
	URL    string
	Branch string
}

type Manager struct {
	cacheRoot string
	config    map[string]CatalogConfig
	db        *gorm.DB
}

func NewManager(cacheRoot string, config map[string]CatalogConfig, db *gorm.DB) *Manager {
	return &Manager{
		cacheRoot: cacheRoot,
		config:    config,
		db:        db,
	}
}

func (m *Manager) RefreshAll() error {
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
		catalog.TransitioningMessage = err.Error()
		return m.updateDb(catalog, nil, nil, commit)
	}

	// Catalog is already up to date
	if commit == catalog.Commit {
		return nil
	}

	templates, versions, errors := traverseFiles(repoPath)
	if errors != nil {
		catalog.TransitioningMessage = strings.Join(errors, ";")
	}

	return m.updateDb(catalog, templates, versions, commit)
}
