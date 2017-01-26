package manager

import (
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/rancher/catalog-service/model"
)

// TODO: move elsewhere
type CatalogConfig struct {
	URL    string
	Branch string
	// TODO: remove this
	EnvironmentId string
}

type Manager struct {
	cacheRoot string
	diskLocks map[string]*sync.Mutex
	//catalogURLs []string
	config map[string]CatalogConfig
	db     *gorm.DB
}

//func NewManager(cacheRoot string, catalogURLs []string) *Manager {
func NewManager(cacheRoot string, config map[string]CatalogConfig, db *gorm.DB) *Manager {
	return &Manager{
		cacheRoot: cacheRoot,
		config:    config,
		db:        db,
	}
}

func (m *Manager) CreateConfigCatalogs() error {
	for name, config := range m.config {
		if err := m.db.Create(&model.CatalogModel{
			Catalog: model.Catalog{
				Name: name,
				URL:  config.URL,
				// TODO
				EnvironmentId: "e1",
			},
		}).Error; err != nil {
			return err
		}
	}
	return nil
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
	repoPath, err := m.prepareRepoPath(catalog)
	if err != nil {
		return err
	}

	templates, versions, err := traverseFiles(repoPath)
	if err != nil {
		return err
	}

	return m.updateDb(catalog, templates, versions)
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
