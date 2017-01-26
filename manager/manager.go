package manager

import (
	"sync"

	"github.com/jinzhu/gorm"
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

func (m *Manager) RefreshAll() error {
	for catalogName, catalogConfig := range m.config {
		if err := m.refreshCatalog(catalogName, catalogConfig); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Refresh(environmentId string) error {
	for catalogName, catalogConfig := range m.config {
		if catalogConfig.EnvironmentId == environmentId {
			if err := m.refreshCatalog(catalogName, catalogConfig); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) refreshCatalog(name string, config CatalogConfig) error {
	repoPath, err := m.prepareRepoPath(name, config)
	if err != nil {
		return err
	}

	templates, versions, err := traverseFiles(repoPath)
	if err != nil {
		return err
	}

	return m.updateDb(name, config, templates, versions)
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
