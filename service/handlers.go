package service

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"github.com/rancher/go-rancher/api"
	"github.com/rancher/go-rancher/client"
)

const (
	environmentIdHeader = "x-api-project-id"
)

func getEnvironmentId(r *http.Request) (string, error) {
	fmt.Println("Environment header: ", r.Header.Get(environmentIdHeader))
	// TODO
	return "e1", nil
	environment := r.Header.Get(environmentIdHeader)
	if environment == "" {
		return "", fmt.Errorf("Request is missing environment header %s", environment)
	}
	return environment, nil
}

func ReturnHTTPError(w http.ResponseWriter, r *http.Request, httpStatus int, err error) {
	w.WriteHeader(httpStatus)

	catalogError := model.CatalogError{
		Resource: client.Resource{
			Type: "error",
		},
		Status:  strconv.Itoa(httpStatus),
		Message: err.Error(),
	}

	api.CreateApiContext(w, r, schemas)
	api.GetApiContext(r).Write(&catalogError)
}

func getCatalogs(w http.ResponseWriter, r *http.Request) {
	apiContext := api.GetApiContext(r)

	environmentId, err := getEnvironmentId(r)
	if err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}

	var catalogs []model.CatalogModel
	db.Find(&catalogs, &model.CatalogModel{
		Catalog: model.Catalog{
			EnvironmentId: environmentId,
		},
	})

	resp := model.CatalogCollection{}
	for _, catalog := range catalogs {
		resp.Data = append(resp.Data, model.CatalogResource{
			// TODO: better id
			Resource: client.Resource{
				Id:   catalog.Name,
				Type: "catalog",
			},
			Catalog: catalog.Catalog,
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
	var catalog model.CatalogModel
	db.Where(&model.CatalogModel{
		Catalog: model.Catalog{
			Name:          catalogName,
			EnvironmentId: environmentId,
		},
	}).First(&catalog)

	apiContext.Write(&model.CatalogResource{
		Resource: client.Resource{
			Id:   catalog.Name,
			Type: "catalog",
		},
		Catalog: catalog.Catalog,
	})
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

	/*templateBaseEq := r.URL.Query().Get("templateBase_eq")
	templateBaseNe := r.URL.Query().Get("templateBase_ne")
	minumumRancherVersionLte := r.URL.Query().Get("minumumRancherVersion_lte")
	maximumRancherVersionGte := r.URL.Query().Get("maximumRancherVersion_gte")*/

	var templates []model.TemplateModel
	query := model.TemplateModel{
		Template: model.Template{
			EnvironmentId: environmentId,
		},
	}
	if catalog != "" {
		query.Catalog = catalog
	}
	if category != "" {
		query.Category = category
	}
	db.Find(&templates, &query)

	resp := model.TemplateCollection{}
	for _, template := range templates {
		// TODO: this is duplicated
		// TODO: shouldn't need to lookup all versions for this
		var versionModels []model.VersionModel
		db.Find(&versionModels, &model.VersionModel{
			Version: model.Version{
				Template:      template.FolderName,
				EnvironmentId: environmentId,
			},
		})

		var versions []model.Version
		for _, versionModel := range versionModels {
			versions = append(versions, versionModel.Version)
		}

		resp.Data = append(resp.Data, *templateResource(apiContext, template.Template, versions))
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

	catalogName, templateName, templateBase, revisionNumber, _ := parse.TemplateURLPath(catalogTemplateVersion)
	if revisionNumber == -1 {
		// Return template
		var templateModel model.TemplateModel
		var versionModels []model.VersionModel
		db.Where(&model.VersionModel{
			Version: model.Version{
				Catalog:       catalogName,
				Template:      templateName,
				EnvironmentId: environmentId,
			},
		}).Find(&versionModels)
		db.Where(&model.TemplateModel{
			Template: model.Template{
				Catalog:       catalogName,
				FolderName:    templateName,
				EnvironmentId: environmentId,
				Base:          templateBase,
			},
		}).First(&templateModel)

		if r.URL.RawQuery != "" && strings.EqualFold("image", r.URL.RawQuery) {
			iconReader := bytes.NewReader(templateModel.Icon)
			http.ServeContent(w, r, templateModel.IconFilename, time.Time{}, iconReader)
			return
		}

		var versions []model.Version
		for _, versionModel := range versionModels {
			versions = append(versions, versionModel.Version)
		}
		apiContext.Write(templateResource(apiContext, templateModel.Template, versions))
	} else {
		// Return template version
		var template model.TemplateModel
		var version model.VersionModel
		var versionModels []model.VersionModel
		db.Where(&model.TemplateModel{
			Template: model.Template{
				Catalog:       catalogName,
				FolderName:    templateName,
				EnvironmentId: environmentId,
				Base:          templateBase,
			},
		}).First(&template)
		db.Where(&model.VersionModel{
			Version: model.Version{
				Catalog:       catalogName,
				Template:      templateName,
				EnvironmentId: environmentId,
			},
		}).Find(&versionModels)
		db.Where(&model.VersionModel{
			Version: model.Version{
				Catalog:       catalogName,
				Template:      templateName,
				EnvironmentId: environmentId,
				Revision:      revisionNumber,
			},
		}).First(&version)

		if r.URL.RawQuery != "" && strings.EqualFold("readme", r.URL.RawQuery) {
			w.Write([]byte(version.Readme))
			return
		}

		var fileModels []model.FileModel
		db.Where(&model.FileModel{
			File: model.File{
				Catalog:       catalogName,
				EnvironmentId: environmentId,
				VersionID:     version.ID,
			},
		}).Find(&fileModels)
		var files []model.File
		for _, fileModel := range fileModels {
			files = append(files, fileModel.File)
		}

		var versions []model.Version
		for _, versionModel := range versionModels {
			versions = append(versions, versionModel.Version)
		}
		versionResource, err := versionResource(apiContext, template.Template, version.Version, versions, files)
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

	var templates []model.TemplateModel
	db.Find(&templates, &model.TemplateModel{
		Template: model.Template{
			Catalog:       catalogName,
			EnvironmentId: environmentId,
		},
	})

	resp := model.TemplateCollection{}
	for _, template := range templates {
		// TODO: this is duplicated
		// TODO: shouldn't need to lookup all versions for this
		var versionModels []model.VersionModel
		db.Find(&versionModels, &model.VersionModel{
			Version: model.Version{
				Template:      template.FolderName,
				EnvironmentId: environmentId,
			},
		})

		var versions []model.Version
		for _, versionModel := range versionModels {
			versions = append(versions, versionModel.Version)
		}

		resp.Data = append(resp.Data, *templateResource(apiContext, template.Template, versions))
	}

	resp.Actions = map[string]string{
		"refresh": api.GetApiContext(r).UrlBuilder.ReferenceByIdLink("template", "") + "?action=refresh",
	}

	apiContext.Write(&resp)

}

func refreshCatalog(w http.ResponseWriter, r *http.Request) {
	if err := m.RefreshAll(); err != nil {
		ReturnHTTPError(w, r, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
