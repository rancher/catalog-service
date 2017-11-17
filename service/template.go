package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	catalogClient "github.com/rancher/catalog-service/client"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"github.com/rancher/go-rancher/api"
	catalogv1 "github.com/rancher/type/apis/catalog.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Removes template that belongs to a duplicate catalog
func removeDuplicateCatalogTemplate(templates []model.Template, envId string) []model.Template {
	catalogNames := []string{}
	templateMap := make(map[string][]model.Template)
	for _, template := range templates {
		if _, exists := templateMap[template.CatalogName]; !exists {
			templateMap[template.CatalogName] = []model.Template{}
		}

		templateMap[template.CatalogName] = append(templateMap[template.CatalogName], template)
		catalogNames = append(catalogNames, template.CatalogName)
	}

	nameSet := map[string]struct{}{}
	for _, name := range catalogNames {
		nameSet[name] = struct{}{}
	}
	notExistingNames := []string{}
	for name := range nameSet {
		_, err := catalogClient.CatalogClient.Get(name, metav1.GetOptions{})
		if err != nil {
			notExistingNames = append(notExistingNames, name)
			continue
		}
	}

	for _, catalogName := range notExistingNames {
		if _, exist := templateMap[catalogName]; exist {
			delete(templateMap, catalogName)
		}
	}

	finalTemplates := []model.Template{}
	for _, templateSlice := range templateMap {
		finalTemplates = append(finalTemplates, templateSlice...)
	}

	return finalTemplates

}

func getTemplates(w http.ResponseWriter, r *http.Request, envId string, catalogClient catalogv1.CatalogInterface) (int, error) {
	apiContext := api.GetApiContext(r)

	catalog := r.URL.Query().Get("catalogId")
	if catalog == "" {
		catalog = r.URL.Query().Get("catalog")
	}
	rancherVersion := r.URL.Query().Get("rancherVersion")

	// Backwards compatibility for older versions of CLI
	minRancherVersion := r.URL.Query().Get("minimumRancherVersion_lte")
	if rancherVersion == "" && minRancherVersion != "" {
		rancherVersion = minRancherVersion
	}

	templateBaseEq := r.URL.Query().Get("templateBase_eq")
	categories, _ := r.URL.Query()["category"]
	categoriesNe, _ := r.URL.Query()["category_ne"]

	templates := model.LookupTemplates(db, envId, catalog, templateBaseEq, categories, categoriesNe)
	templates = removeDuplicateCatalogTemplate(templates, envId)

	resp := model.TemplateCollection{}
	for _, template := range templates {
		templateResource := templateResource(apiContext, template.Catalog, template, rancherVersion, envId)
		if len(templateResource.VersionLinks) > 0 {
			resp.Data = append(resp.Data, *templateResource)
		}
	}

	resp.Actions = map[string]string{
		"refresh": api.GetApiContext(r).UrlBuilder.ReferenceByIdLink("template", "") + "?action=refresh",
	}

	apiContext.Write(&resp)
	return 0, nil
}

func getTemplate(w http.ResponseWriter, r *http.Request, envId string, catalogClient catalogv1.CatalogInterface) (int, error) {
	apiContext := api.GetApiContext(r)
	vars := mux.Vars(r)

	catalogTemplateVersion, ok := vars["catalog_template_version"]
	if !ok {
		return http.StatusBadRequest, errors.New("Missing paramater catalog_template_version")
	}

	rancherVersion := r.URL.Query().Get("rancherVersion")

	catalogName, templateName, templateBase, revisionOrVersion, _ := parse.TemplateURLPath(catalogTemplateVersion)

	template := model.LookupTemplate(db, envId, catalogName, templateName, templateBase)

	if template == nil {
		return http.StatusNotFound, errors.New("Template not found")
	}

	if revisionOrVersion == "" {
		if _, ok := r.URL.Query()["image"]; ok {
			icon, err := base64.StdEncoding.DecodeString(template.Icon)
			if err != nil {
				return http.StatusBadRequest, err
			}
			iconReader := bytes.NewReader(icon)
			http.ServeContent(w, r, template.IconFilename, time.Time{}, iconReader)
			return 0, nil
		} else if r.URL.RawQuery != "" && strings.EqualFold("readme", r.URL.RawQuery) {
			w.Write([]byte(template.Readme))
			return 0, nil
		}

		// Return template
		apiContext.Write(templateResource(apiContext, catalogName, *template, rancherVersion, envId))
	} else {
		var version *model.Version
		revision, err := strconv.Atoi(revisionOrVersion)
		if err == nil {
			version = model.LookupVersionByRevision(db, envId, catalogName, templateBase, templateName, revision)
		} else {
			version = model.LookupVersionByVersion(db, envId, catalogName, templateBase, templateName, revisionOrVersion)
		}
		if version == nil {
			return http.StatusNotFound, errors.New("Version not found")
		}

		if r.URL.RawQuery != "" && strings.EqualFold("readme", r.URL.RawQuery) {
			w.Write([]byte(version.Readme))
			return 0, nil
		}

		versionResource, err := versionResource(apiContext, catalogName, *template, *version, rancherVersion, envId)
		if err != nil {
			return http.StatusBadRequest, err
		}

		// Return template version
		apiContext.Write(versionResource)
	}

	return 0, nil
}

func refreshTemplates(w http.ResponseWriter, r *http.Request, envId string, catalogClient catalogv1.CatalogInterface) (int, error) {
	if err := m.Refresh(envId, true); err != nil {
		return http.StatusInternalServerError, err
	}
	if envId != "global" {
		if err := m.Refresh("global", true); err != nil {
			return http.StatusInternalServerError, err
		}
	}
	w.WriteHeader(http.StatusNoContent)
	return 0, nil
}
