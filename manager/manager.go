package manager

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/rancher/catalog-service/model"
	catalogv1 "github.com/rancher/type/apis/catalog.cattle.io/v1"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	HelmTemplateType     = "helm"
	RancherTemplateType  = "native"
	HelmTemplateBaseType = "kubernetes"
)

type Manager struct {
	cacheRoot     string
	config        map[string]CatalogConfig
	strict        bool
	db            *gorm.DB
	uuid          string
	httpClient    http.Client
	catalogClient catalogv1.CatalogInterface
}

func NewManager(cacheRoot string, strict bool, db *gorm.DB, uuid string, catalogClient catalogv1.CatalogInterface) *Manager {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	return &Manager{
		cacheRoot:     cacheRoot,
		strict:        strict,
		db:            db,
		uuid:          uuid,
		httpClient:    client,
		catalogClient: catalogClient,
	}
}

func (m *Manager) HandleCatalog(key string, catalog *catalogv1.Catalog) error {
	_, err := m.catalogClient.Get(key, metav1.GetOptions{})
	if err == nil {
		log.Debugf("refreshing catalog %s by controller", key)
		return m.refreshCatalog(*catalog, true)
	} else if !kerrors.IsNotFound(err) {
		return err
	}
	delete(m.config, key)
	return nil
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

type RepoRefreshError struct {
	Errors []error
}

func (e *RepoRefreshError) Error() string {
	return fmt.Sprintf("%v", e.Errors)
}

func (m *Manager) refreshConfigCatalogs(update bool) error {
	catalogList, err := m.catalogClient.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	catalogConfig := map[string]CatalogConfig{}
	for _, catalog := range catalogList.Items {
		catalogConfig[catalog.Name] = CatalogConfig{
			URL:    catalog.Spec.URL,
			Branch: catalog.Spec.Branch,
			Kind:   catalog.Spec.CatalogKind,
		}
	}
	m.config = catalogConfig

	var refreshErrors []error
	for name, config := range m.config {
		catalog := catalogv1.Catalog{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					model.ProjectLabel: "global",
				},
			},
			TypeMeta: metav1.TypeMeta{
				APIVersion: "catalog.cattle.io/v1",
				Kind:       "Catalog",
			},
			Spec: catalogv1.CatalogSpec{
				URL:         config.URL,
				Branch:      config.Branch,
				CatalogKind: config.Kind,
			},
		}
		existingCatalog, err := m.lookupCatalog("global", name)
		if err == nil && existingCatalog.Spec.URL == catalog.Spec.URL && existingCatalog.Spec.Branch == catalog.Spec.Branch {
			catalog = existingCatalog
		}
		if err := m.refreshCatalog(catalog, update); err != nil {
			refreshErrors = append(refreshErrors, errors.Wrapf(err, "Catalog refresh failed for %v (%v)", catalog.Name, catalog.Spec.URL))
		}
	}
	if len(refreshErrors) > 0 {
		return &RepoRefreshError{Errors: refreshErrors}
	}
	return nil
}

func (m *Manager) refreshEnvironmentCatalogs(environmentId string, update bool) error {
	catalogs, err := m.lookupCatalogs(environmentId)
	if err != nil {
		return err
	}

	var refreshErrors []error
	for _, catalog := range catalogs {
		if err := m.refreshCatalog(catalog, update); err != nil {
			refreshErrors = append(refreshErrors, errors.Wrapf(err, "Catalog refresh failed for %v (%v)", catalog.Name, catalog.Spec.URL))
		}
	}
	if len(refreshErrors) > 0 {
		return &RepoRefreshError{Errors: refreshErrors}
	}
	return nil
}

func (m *Manager) refreshCatalog(catalog catalogv1.Catalog, update bool) error {
	repoPath, commit, catalogType, err := m.prepareRepoPath(catalog, update)
	if err != nil {
		return err
	}

	// Catalog is already up to date
	if commit == catalog.Spec.Commit {
		log.Debugf("Catalog %s is already up to date", catalog.Name)
		return nil
	}

	templates, errs, err := traverseFiles(repoPath, catalog.Spec.CatalogKind, catalogType)
	if err != nil {
		return errors.Wrap(err, "Repo traversal failed")
	}

	if len(errs) != 0 {
		if m.strict {
			return fmt.Errorf("%v", errs)
		}
		log.Errorf("Errors while parsing repo: %v", errs)
	}

	log.Debugf("Updating catalog %s", catalog.Name)
	return m.updateDb(catalog, templates, commit)
}
