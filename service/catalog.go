package service

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/go-rancher/api"
	"github.com/rancher/go-rancher/client"
)

func getCatalogs(w http.ResponseWriter, r *http.Request, envId string) error {
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
	return nil
}

func getCatalog(w http.ResponseWriter, r *http.Request, envId string) error {
	apiContext := api.GetApiContext(r)

	vars := mux.Vars(r)
	envId, err := getEnvironmentId(r)
	if err != nil {
		return err
	}

	// TODO error checking
	catalogName := vars["catalog"]
	catalog := model.LookupCatalog(db, envId, catalogName)

	apiContext.Write(catalogResource(*catalog))
	return nil
}

func createCatalog(w http.ResponseWriter, r *http.Request, envId string) error {
	apiContext := api.GetApiContext(r)

	catalogName := r.FormValue("name")
	url := r.FormValue("url")
	branch := r.FormValue("branch")

	if catalogName == "" {
		return errors.New("Missing field 'name'")
	}
	if url == "" {
		return errors.New("Missing field 'url'")
	}

	catalogModel := model.CatalogModel{
		Catalog: model.Catalog{
			EnvironmentId: envId,
			Name:          catalogName,
			URL:           url,
			Branch:        branch,
		},
	}

	if err := db.Create(&catalogModel).Error; err != nil {
		return err
	}

	apiContext.Write(catalogResource(catalogModel.Catalog))
	return nil
}

func deleteCatalog(w http.ResponseWriter, r *http.Request, envId string) error {
	vars := mux.Vars(r)

	// TODO error checking
	name := vars["catalog"]

	model.DeleteCatalog(db, envId, name)

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func getCatalogTemplates(w http.ResponseWriter, r *http.Request, envId string) error {
	apiContext := api.GetApiContext(r)
	vars := mux.Vars(r)

	catalogName, ok := vars["catalog"]
	if !ok {
		return errors.New("Missing paramater catalog")
	}

	category := r.URL.Query().Get("category")
	//categoryNe := r.URL.Query().Get("category_ne")
	rancherVersion := r.URL.Query().Get("rancherVersion")

	templates := model.LookupCatalogTemplates(db, envId, catalogName, category)

	// TODO: this is duplicated
	resp := model.TemplateCollection{}
	for _, template := range templates {
		resp.Data = append(resp.Data, *templateResource(apiContext, catalogName, template, rancherVersion))
	}

	resp.Actions = map[string]string{
		"refresh": api.GetApiContext(r).UrlBuilder.ReferenceByIdLink("template", "") + "?action=refresh",
	}

	apiContext.Write(&resp)
	return nil
}
