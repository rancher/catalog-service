package service

import (
	"bytes"
	"encoding/base64"
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
	rancherVersion := r.URL.Query().Get("rancherVersion")
	templateBaseEq := r.URL.Query().Get("templateBase_eq")
	categories, _ := r.URL.Query()["category"]
	categoriesNe, _ := r.URL.Query()["category_ne"]

	templates := model.LookupTemplates(db, envId, catalog, templateBaseEq, categories, categoriesNe)

	resp := model.TemplateCollection{}
	for _, template := range templates {
		catalog := model.GetCatalog(db, template.CatalogId)
		resp.Data = append(resp.Data, *templateResource(apiContext, catalog.Name, template, rancherVersion))
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
			icon, err := base64.StdEncoding.DecodeString(template.Icon)
			if err != nil {
				return nil
			}
			iconReader := bytes.NewReader(icon)
			http.ServeContent(w, r, template.IconFilename, time.Time{}, iconReader)
			return nil
		} else if r.URL.RawQuery != "" && strings.EqualFold("readme", r.URL.RawQuery) {
			w.Write([]byte(template.Readme))
			return nil
		}

		apiContext.Write(templateResource(apiContext, catalogName, *template, rancherVersion))
	} else {
		// Return template version
		template := model.LookupTemplate(db, envId, catalogName, templateName, templateBase)
		version := model.LookupVersion(db, envId, catalogName, templateBase, templateName, revisionNumber)

		if r.URL.RawQuery != "" && strings.EqualFold("readme", r.URL.RawQuery) {
			w.Write([]byte(version.Readme))
			return nil
		}

		versionResource, err := versionResource(apiContext, catalogName, *template, *version, rancherVersion)
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
	if envId != "global" {
		if err := m.Refresh("global"); err != nil {
			return err
		}
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
