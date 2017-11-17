package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/go-rancher/api"
	catalogv1 "github.com/rancher/type/apis/catalog.cattle.io/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getCatalogs(w http.ResponseWriter, r *http.Request, envID string, catalogClient catalogv1.CatalogInterface) (int, error) {
	apiContext := api.GetApiContext(r)

	catalogList, err := catalogClient.List(metav1.ListOptions{})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	resp := model.CatalogCollection{}
	for _, catalog := range catalogList.Items {
		resp.Data = append(resp.Data, *catalogResource(catalog, apiContext, envID))
	}

	apiContext.Write(&resp)
	return 0, nil
}

func getCatalog(w http.ResponseWriter, r *http.Request, envID string, catalogClient catalogv1.CatalogInterface) (int, error) {
	apiContext := api.GetApiContext(r)

	vars := mux.Vars(r)
	envID, err := getEnvironmentId(r)
	if err != nil {
		return http.StatusBadRequest, err
	}
	// TODO error checking
	catalogName := vars["catalog"]
	catalog, err := catalogClient.Get(catalogName, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		return http.StatusNotFound, errors.New("Catalog not found")
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	apiContext.Write(catalogResource(*catalog, apiContext, envID))
	return 0, nil
}

type CatalogRequest struct {
	Name   string
	URL    string
	Branch string
	Kind   string
}

func isDuplicateName(catalog *catalogv1.Catalog, catalogClient catalogv1.CatalogInterface) (bool, error) {
	_, err := catalogClient.Get(catalog.Name, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func isDuplicateEnvName(catalog *catalogv1.Catalog, oldCatalogName string, catalogClient catalogv1.CatalogInterface) (bool, error) {
	if oldCatalogName == catalog.Name {
		return false, nil
	}

	_, err := catalogClient.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", model.ProjectLabel, catalog.Labels[model.ProjectLabel]),
	})
	if kerrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func isDuplicateGlobalName(catalog *catalogv1.Catalog, catalogClient catalogv1.CatalogInterface) (bool, error) {
	_, err := catalogClient.Get(catalog.Name, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func createCatalog(w http.ResponseWriter, r *http.Request, envId string, catalogClient catalogv1.CatalogInterface) (int, error) {
	apiContext := api.GetApiContext(r)

	catalogModel, err := catalogModelFromRequest(r, envId)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if catalogModel.Name == "" {
		return http.StatusBadRequest, errors.New("Missing field 'name'")
	}
	if catalogModel.Spec.URL == "" {
		return http.StatusBadRequest, errors.New("Missing field 'url'")
	}

	if exist, err := isDuplicateName(&catalogModel.Catalog, catalogClient); err == nil && exist {
		return http.StatusUnprocessableEntity, fmt.Errorf("Catalog name %s already exists", catalogModel.Name)
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	catalog, err := catalogClient.Create(&catalogModel.Catalog)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	apiContext.Write(catalogResource(*catalog, apiContext, ""))
	return 0, nil
}

func catalogExists(catalog *catalogv1.Catalog, catalogClient catalogv1.CatalogInterface) (bool, error) {
	_, err := catalogClient.Get(catalog.Name, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func updateCatalog(w http.ResponseWriter, r *http.Request, envId string, catalogClient catalogv1.CatalogInterface) (int, error) {
	apiContext := api.GetApiContext(r)

	catalogModel, err := catalogModelFromRequest(r, envId)
	if err != nil {
		return http.StatusBadRequest, err
	}

	vars := mux.Vars(r)
	oldCatalogName := vars["catalog"]

	oldCatalog := model.CatalogModel{
		Catalog: catalogv1.Catalog{
			ObjectMeta: metav1.ObjectMeta{
				Name: oldCatalogName,
				Labels: map[string]string{
					model.ProjectLabel: envId,
				},
			},
		},
	}

	if exist, err := isDuplicateGlobalName(&catalogModel.Catalog, catalogClient); err == nil && exist {
		return http.StatusUnprocessableEntity, fmt.Errorf("Catalog name %s already exists", catalogModel.Name)
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	if exist, err := isDuplicateEnvName(&catalogModel.Catalog, oldCatalogName, catalogClient); err == nil && exist {
		return http.StatusUnprocessableEntity, fmt.Errorf("Catalog name %s already exists", catalogModel.Name)
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	if exist, err := catalogExists(&oldCatalog.Catalog, catalogClient); err == nil && !exist {
		return http.StatusNotFound, errors.New("Catalog not found")
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	catalog, err := catalogClient.Update(&catalogModel.Catalog)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	apiContext.Write(catalogResource(*catalog, apiContext, ""))
	return 0, nil
}

func catalogModelFromRequest(r *http.Request, envId string) (*model.CatalogModel, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var catalogRequest CatalogRequest
	if err := json.Unmarshal(body, &catalogRequest); err != nil {
		return nil, err
	}

	name := catalogRequest.Name
	if envId != "global" {
		name = envId + "-" + name
	}
	return &model.CatalogModel{
		Catalog: catalogv1.Catalog{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "catalog.cattle.io/v1",
				Kind:       "Catalog",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					model.ProjectLabel: envId,
				},
			},
			Spec: catalogv1.CatalogSpec{
				URL:         catalogRequest.URL,
				Branch:      catalogRequest.Branch,
				CatalogKind: catalogRequest.Kind,
			},
		},
	}, nil
}

func deleteCatalog(w http.ResponseWriter, r *http.Request, envId string, catalogClient catalogv1.CatalogInterface) (int, error) {
	vars := mux.Vars(r)

	name, ok := vars["catalog"]
	if !ok {
		return http.StatusBadRequest, errors.New("Missing paramater catalog")
	}

	if err := catalogClient.Delete(name, &metav1.DeleteOptions{}); err != nil {
		return http.StatusInternalServerError, err
	}

	w.WriteHeader(http.StatusNoContent)
	return 0, nil
}

func getCatalogTemplates(w http.ResponseWriter, r *http.Request, envId string, catalogClient catalogv1.CatalogInterface) (int, error) {
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
		templateResource := templateResource(apiContext, catalogName, template, rancherVersion, envId)
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
