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
	"github.com/rancher/go-rancher/client"
)

func getCatalogs(w http.ResponseWriter, r *http.Request) {
	apiContext := api.GetApiContext(r)

	environmentId, err := getEnvironmentId(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}

	catalogs := model.LookupCatalogs(db, environmentId)

	resp := model.CatalogCollection{}
	for _, catalog := range catalogs {
		resp.Data = append(resp.Data, model.CatalogResource{
			Resource: client.Resource{
				Id:   catalog.Name,
				Type: "catalog",
			},
			Catalog: catalog,
		})
	}

	apiContext.Write(&resp)
}

func getCatalog(w http.ResponseWriter, r *http.Request) {
	apiContext := api.GetApiContext(r)

	vars := mux.Vars(r)
	environmentId, err := getEnvironmentId(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}

	// TODO error checking
	catalogName := vars["catalog"]
	catalog := model.LookupCatalog(db, environmentId, catalogName)

	apiContext.Write(catalogResource(*catalog))
}

func createCatalog(w http.ResponseWriter, r *http.Request) {
	apiContext := api.GetApiContext(r)

	environmentId, err := getEnvironmentId(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}

	catalogName := r.FormValue("name")
	url := r.FormValue("url")
	branch := r.FormValue("branch")

	if catalogName == "" {
		ReturnHTTPError(w, r, http.StatusBadRequest, errors.New("Missing field 'name'"))
		return
	}
	if url == "" {
		ReturnHTTPError(w, r, http.StatusBadRequest, errors.New("Missing field 'url'"))
		return
	}

	catalogModel := model.CatalogModel{
		Catalog: model.Catalog{
			EnvironmentId: environmentId,
			Name:          catalogName,
			URL:           url,
			Branch:        branch,
		},
	}

	if err := db.Create(&catalogModel).Error; err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}

	apiContext.Write(catalogResource(catalogModel.Catalog))
}

func deleteCatalog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	environmentId, err := getEnvironmentId(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}

	// TODO error checking
	name := vars["catalog"]

	model.DeleteCatalog(db, environmentId, name)

	w.WriteHeader(http.StatusNoContent)
}

func getTemplates(w http.ResponseWriter, r *http.Request) {
	apiContext := api.GetApiContext(r)

	environmentId, err := getEnvironmentId(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}

	catalog := r.URL.Query().Get("catalogId")
	if catalog == "" {
		catalog = r.URL.Query().Get("catalog")
	}
	category := r.URL.Query().Get("category")
	//categoryNe := r.URL.Query().Get("category_ne")
	rancherVersion := r.URL.Query().Get("rancherVersion")

	templates := model.LookupTemplates(db, environmentId, catalog, category)

	resp := model.TemplateCollection{}
	for _, template := range templates {
		versions := model.LookupVersions(db, environmentId, catalog, template.FolderName)
		resp.Data = append(resp.Data, *templateResource(apiContext, template, versions, rancherVersion))
	}

	resp.Actions = map[string]string{
		"refresh": api.GetApiContext(r).UrlBuilder.ReferenceByIdLink("template", "") + "?action=refresh",
	}

	apiContext.Write(&resp)
}

func getTemplate(w http.ResponseWriter, r *http.Request) {
	apiContext := api.GetApiContext(r)
	vars := mux.Vars(r)

	environmentId, err := getEnvironmentId(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}

	catalogTemplateVersion, ok := vars["catalog_template_version"]
	if !ok {
		ReturnHTTPError(w, r, http.StatusBadRequest, errors.New("Missing paramater catalog_template_version"))
		return
	}

	rancherVersion := r.URL.Query().Get("rancherVersion")

	catalogName, templateName, templateBase, revisionNumber, _ := parse.TemplateURLPath(catalogTemplateVersion)
	if revisionNumber == -1 {
		// Return template
		template := model.LookupTemplate(db, environmentId, catalogName, templateName, templateBase)

		if r.URL.RawQuery != "" && strings.EqualFold("image", r.URL.RawQuery) {
			iconReader := bytes.NewReader(template.Icon)
			http.ServeContent(w, r, template.IconFilename, time.Time{}, iconReader)
			return
		}

		versions := model.LookupVersions(db, environmentId, catalogName, templateName)

		apiContext.Write(templateResource(apiContext, *template, versions, rancherVersion))
	} else {
		// Return template version
		template := model.LookupTemplate(db, environmentId, catalogName, templateName, templateBase)
		versionModel := model.LookupVersionModel(db, environmentId, catalogName, templateName, revisionNumber)
		versions := model.LookupVersions(db, environmentId, catalogName, templateName)

		// TODO: version READMEs
		if r.URL.RawQuery != "" && strings.EqualFold("readme", r.URL.RawQuery) {
			w.Write([]byte(versionModel.Readme))
			return
		}

		files := model.LookupFiles(db, environmentId, catalogName, versionModel.ID)

		versionResource, err := versionResource(apiContext, *template, versionModel.Version, versions, files, rancherVersion)
		if err != nil {
			ReturnHTTPError(w, r, http.StatusBadRequest, err)
			return
		}
		apiContext.Write(versionResource)
	}
}

func getCatalogTemplates(w http.ResponseWriter, r *http.Request) {
	apiContext := api.GetApiContext(r)
	vars := mux.Vars(r)

	environmentId, err := getEnvironmentId(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}

	catalogName, ok := vars["catalog"]
	if !ok {
		ReturnHTTPError(w, r, http.StatusBadRequest, errors.New("Missing paramater catalog"))
		return
	}

	category := r.URL.Query().Get("category")
	//categoryNe := r.URL.Query().Get("category_ne")
	rancherVersion := r.URL.Query().Get("rancherVersion")

	templates := model.LookupTemplates(db, environmentId, catalogName, category)

	// TODO: this is duplicated
	resp := model.TemplateCollection{}
	for _, template := range templates {
		versions := model.LookupVersions(db, environmentId, catalogName, template.FolderName)
		resp.Data = append(resp.Data, *templateResource(apiContext, template, versions, rancherVersion))
	}

	resp.Actions = map[string]string{
		"refresh": api.GetApiContext(r).UrlBuilder.ReferenceByIdLink("template", "") + "?action=refresh",
	}

	apiContext.Write(&resp)

}

func refreshCatalog(w http.ResponseWriter, r *http.Request) {
	environmentId, err := getEnvironmentId(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}
	if err := m.Refresh(environmentId); err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
