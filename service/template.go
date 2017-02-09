package service

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"github.com/rancher/go-rancher/api"
)

func getTemplates(w http.ResponseWriter, r *http.Request, envId string) error {
	apiContext := api.GetApiContext(r)

	catalog := r.URL.Query().Get("catalogId")
	if catalog == "" {
		catalog = r.URL.Query().Get("catalog")
	}
	category := r.URL.Query().Get("category")
	//categoryNe := r.URL.Query().Get("category_ne")
	rancherVersion := r.URL.Query().Get("rancherVersion")

	var templates []model.Template
	if catalog == "" {
		templates = model.LookupTemplates(db, envId, category)
	} else {
		templates = model.LookupCatalogTemplates(db, envId, catalog, category)
	}

	resp := model.TemplateCollection{}
	for _, template := range templates {
		catalog := model.GetCatalog(db, template.CatalogId)
		versions := model.LookupVersions(db, envId, catalog.Name, template.FolderName)
		resp.Data = append(resp.Data, *templateResource(apiContext, catalog.Name, template, versions, rancherVersion))
	}

	resp.Actions = map[string]string{
		"refresh": api.GetApiContext(r).UrlBuilder.ReferenceByIdLink("template", "") + "?action=refresh",
	}

	apiContext.Write(&resp)
	return nil
}

func getTemplate(w http.ResponseWriter, r *http.Request, envId string) error {
	apiContext := api.GetApiContext(r)
	vars := mux.Vars(r)

	catalogTemplateVersion, ok := vars["catalog_template_version"]
	if !ok {
		return errors.New("Missing paramater catalog_template_version")
	}

	rancherVersion := r.URL.Query().Get("rancherVersion")

	catalogName, templateName, templateBase, revisionNumber, _ := parse.TemplateURLPath(catalogTemplateVersion)
	if revisionNumber == -1 {
		// Return template
		template := model.LookupTemplate(db, envId, catalogName, templateName, templateBase)

		if r.URL.RawQuery != "" && strings.EqualFold("image", r.URL.RawQuery) {
			iconReader := bytes.NewReader(template.Icon)
			http.ServeContent(w, r, template.IconFilename, time.Time{}, iconReader)
			return nil
		}

		versions := model.LookupVersions(db, envId, catalogName, templateName)

		apiContext.Write(templateResource(apiContext, catalogName, *template, versions, rancherVersion))
	} else {
		// Return template version
		template := model.LookupTemplate(db, envId, catalogName, templateName, templateBase)
		versionModel := model.LookupVersionModel(db, envId, catalogName, templateName, revisionNumber)
		versions := model.LookupVersions(db, envId, catalogName, templateName)

		// TODO: version READMEs
		if r.URL.RawQuery != "" && strings.EqualFold("readme", r.URL.RawQuery) {
			w.Write([]byte(versionModel.Readme))
			return nil
		}

		files := model.LookupFiles(db, envId, catalogName, versionModel.ID)

		versionResource, err := versionResource(apiContext, catalogName, *template, versionModel.Version, versions, files, rancherVersion)
		if err != nil {
			return err
		}
		apiContext.Write(versionResource)
	}

	return nil
}

func refreshTemplates(w http.ResponseWriter, r *http.Request, envId string) error {
	if err := m.Refresh(envId); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
