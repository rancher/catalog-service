package manager

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/rancher/catalog-service/model"
)

const (
	HelmTemplateType     = "helm"
	RancherTemplateType  = "native"
	HelmTemplateBaseType = "kubernetes"
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

func (m *Manager) RefreshAll(update bool) error {
	if err := m.refreshConfigCatalogs(update); err != nil {
		return err
	}
	return m.refreshEnvironmentCatalogs("", update)
}

func (m *Manager) Refresh(environmentId string, update bool) error {
	if environmentId == "global" {
		return m.refreshConfigCatalogs(update)
	}
	return m.refreshEnvironmentCatalogs(environmentId, update)
}

func (m *Manager) refreshConfigCatalogs(update bool) error {
	if err := m.readConfig(); err != nil {
		return err
	}
	if err := m.removeCatalogsNotInConfig(); err != nil {
		return err
	}

	var refreshErrors []string
	for name, config := range m.config {
		catalog := model.Catalog{
			Name:          name,
			URL:           config.URL,
			Branch:        config.Branch,
			EnvironmentId: "global",
			Kind:          config.Kind,
		}
		existingCatalog, err := m.lookupCatalog("global", name)
		if err == nil && existingCatalog.URL == catalog.URL && existingCatalog.Branch == catalog.Branch {
			catalog = existingCatalog
		}
		if err := m.refreshCatalog(catalog, update); err != nil {
			refreshErrors = append(refreshErrors, fmt.Sprintf("%s: %v", catalog.Name, err))
		}
	}
	if len(refreshErrors) > 0 {
		return errors.New(strings.Join(refreshErrors, "\n"))
	}
	return nil
}

func (m *Manager) refreshEnvironmentCatalogs(environmentId string, update bool) error {
	catalogs, err := m.lookupCatalogs(environmentId)
	if err != nil {
		return err
	}

	var refreshErrors []string
	for _, catalog := range catalogs {
		if err := m.refreshCatalog(catalog, update); err != nil {
			refreshErrors = append(refreshErrors, fmt.Sprintf("%s: %v", catalog.Name, err))
		}
	}
	if len(refreshErrors) > 0 {
		return errors.New(strings.Join(refreshErrors, "\n"))
	}
	return nil
}

func (m *Manager) refreshCatalog(catalog model.Catalog, update bool) error {
	repoPath, commit, catalogType, err := m.prepareRepoPath(catalog, update)
	if err != nil {
		return err
	}

	// Catalog is already up to date
	if commit == catalog.Commit {
		log.Debugf("Catalog %s is already up to date", catalog.Name)
		return nil
	}

	templates, errors, err := traverseFiles(repoPath, catalog.Kind, catalogType)
	if err != nil {
		return err
	}
	if len(errors) != 0 {
		if m.strict {
			return fmt.Errorf("%v", errors)
		}
		log.Errorf("Errors while parsing repo: %v", errors)
	}

	log.Debugf("Updating catalog %s", catalog.Name)
	return m.updateDb(catalog, templates, commit)
}
