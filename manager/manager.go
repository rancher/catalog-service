package manager

import (
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
		return nil
	}

	templates, err := traverseFiles(repoPath)
	if err != nil {
		return err
	}

	return m.updateDb(catalog, templates, commit)
}

// TODO: move elsewhere
type TemplateConfig struct {
	Name           string            `yaml:"name"`
	Category       string            `yaml:"category"`
	Description    string            `yaml:"description"`
	Version        string            `yaml:"version"`
	Maintainer     string            `yaml:"maintainer"`
	License        string            `yaml:"license"`
	ProjectURL     string            `yaml:"projectURL"`
	IsSystem       string            `yaml:"isSystem"`
	DefaultVersion string            `yaml:"version"`
	Labels         map[string]string `yaml:"version"`
}
