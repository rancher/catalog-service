package service

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/go-rancher/api"
	"github.com/rancher/go-rancher/client"
)

func getCatalogs(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	apiContext := api.GetApiContext(r)

	catalogs := model.LookupCatalogs(db, envId)

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
	return 0, nil
}

func getCatalog(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	apiContext := api.GetApiContext(r)

	vars := mux.Vars(r)
	envId, err := getEnvironmentId(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// TODO error checking
	catalogName := vars["catalog"]
	catalog := model.LookupCatalog(db, envId, catalogName)
	if catalog == nil {
		return http.StatusNotFound, errors.New("Catalog not found")
	}

	apiContext.Write(catalogResource(*catalog))
	return 0, nil
}

type CreateCatalogRequest struct {
	Name   string
	URL    string
	Branch string
}

func createCatalog(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	apiContext := api.GetApiContext(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	var createCatalogRequest CreateCatalogRequest
	if err := json.Unmarshal(body, &createCatalogRequest); err != nil {
		return http.StatusBadRequest, err
	}

	if createCatalogRequest.Name == "" {
		return http.StatusBadRequest, errors.New("Missing field 'name'")
	}
	if createCatalogRequest.URL == "" {
		return http.StatusBadRequest, errors.New("Missing field 'url'")
	}

	catalogModel := model.CatalogModel{
		Catalog: model.Catalog{
			EnvironmentId: envId,
			Name:          createCatalogRequest.Name,
			URL:           createCatalogRequest.URL,
			Branch:        createCatalogRequest.Branch,
		},
	}

	if err := db.Create(&catalogModel).Error; err != nil {
		return http.StatusBadRequest, err
	}

	apiContext.Write(catalogResource(catalogModel.Catalog))
	return 0, nil
}

func deleteCatalog(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	vars := mux.Vars(r)

	name, ok := vars["catalog"]
	if !ok {
		return http.StatusBadRequest, errors.New("Missing paramater catalog")
	}

	model.DeleteCatalog(db, envId, name)

	w.WriteHeader(http.StatusNoContent)
	return 0, nil
}

func getCatalogTemplates(w http.ResponseWriter, r *http.Request, envId string) (int, error) {
	apiContext := api.GetApiContext(r)
	vars := mux.Vars(r)

	catalogName, ok := vars["catalog"]
	if !ok {
		return http.StatusBadRequest, errors.New("Missing paramater catalog")
	}

	rancherVersion := r.URL.Query().Get("rancherVersion")
	templateBaseEq := r.URL.Query().Get("templateBase_eq")
	categories, _ := r.URL.Query()["category"]
	categoriesNe, _ := r.URL.Query()["category_ne"]

	templates := model.LookupTemplates(db, envId, catalogName, templateBaseEq, categories, categoriesNe)

	// TODO: this is duplicated
	resp := model.TemplateCollection{}
	for _, template := range templates {
		templateResource := templateResource(apiContext, catalogName, template, rancherVersion)
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
