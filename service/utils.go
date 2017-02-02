package service

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"github.com/rancher/catalog-service/utils"
	"github.com/rancher/go-rancher/api"
	"github.com/rancher/go-rancher/client"
)

const (
	environmentIdHeader = "x-api-project-id"
)

func getEnvironmentId(r *http.Request) (string, error) {
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

	api.GetApiContext(r).Write(&catalogError)
}

// TODO: this should return an error
func URLEncoded(str string) string {
	u, err := url.Parse(str)
	if err != nil {
		log.Errorf("Error encoding the url: %s , error: %v", str, err)
		return str
	}
	return u.String()
}

func generateVersionId(template model.Template, version model.Version) string {
	if template.Base == "" {
		return fmt.Sprintf("%s:%s:%d", template.Catalog, template.FolderName, version.Revision)
	}
	return fmt.Sprintf("%s:%s*%s:%d", template.Catalog, template.Base, template.FolderName, version.Revision)
}

func generateTemplateId(template model.Template) string {
	if template.Base == "" {
		return fmt.Sprintf("%s:%s", template.Catalog, template.FolderName)
	}
	return fmt.Sprintf("%s:%s*%s", template.Catalog, template.Base, template.FolderName)
}

func catalogResource(catalog model.Catalog) *model.CatalogResource {
	return &model.CatalogResource{
		Resource: client.Resource{
			Id:   catalog.Name,
			Type: "catalog",
		},
		Catalog: catalog,
	}
}

func templateResource(apiContext *api.ApiContext, template model.Template, versions []model.Version, rancherVersion string) *model.TemplateResource {
	templateId := generateTemplateId(template)

	versionLinks := map[string]string{}
	for _, version := range versions {
		if utils.VersionBetween(version.MinimumRancherVersion, rancherVersion, version.MaximumRancherVersion) {
			route := generateVersionId(template, version)
			link := apiContext.UrlBuilder.ReferenceByIdLink("template", route)
			versionLinks[version.Version] = URLEncoded(link)
		}
	}

	links := map[string]string{}
	links["icon"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?image", templateId)))

	return &model.TemplateResource{
		Resource: client.Resource{
			Id:    templateId,
			Type:  "template",
			Links: links,
		},
		Template:     template,
		VersionLinks: versionLinks,
	}
}

func versionResource(apiContext *api.ApiContext, template model.Template, version model.Version, versions []model.Version, files []model.File, rancherVersion string) (*model.TemplateVersionResource, error) {
	templateId := generateTemplateId(template)
	versionId := generateVersionId(template, version)

	filesMap := map[string]string{}
	for _, file := range files {
		filesMap[file.Name] = file.Contents
	}

	var bindings map[string]model.Bindings
	dockerCompose, ok := filesMap["docker-compose.yml"]
	if ok {
		var err error
		bindings, err = parse.Bindings([]byte(dockerCompose))
		if err != nil {
			return nil, err
		}
	}

	var questions []model.Question
	rancherCompose, ok := filesMap["rancher-compose.yml"]
	if ok {
		catalogInfo, err := parse.CatalogInfoFromRancherCompose([]byte(rancherCompose))
		if err != nil {
			return nil, err
		}
		questions = catalogInfo.Questions
	}

	links := map[string]string{}
	links["icon"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?image", templateId)))
	links["readme"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?readme", versionId)))

	upgradeVersionLinks := map[string]string{}
	for _, upgradeVersion := range versions {
		if showUpgradeVersion(version, upgradeVersion, rancherVersion) {
			route := generateVersionId(template, upgradeVersion)
			link := apiContext.UrlBuilder.ReferenceByIdLink("template", route)
			upgradeVersionLinks[upgradeVersion.Version] = URLEncoded(link)
		}
	}

	return &model.TemplateVersionResource{
		Resource: client.Resource{
			Id:    versionId,
			Type:  "templateVersion",
			Links: links,
		},
		Version:             version,
		Bindings:            bindings,
		Files:               filesMap,
		Questions:           questions,
		UpgradeVersionLinks: upgradeVersionLinks,
	}, nil
}

func showUpgradeVersion(version, upgradeVersion model.Version, rancherVersion string) bool {
	if !utils.VersionGreaterThan(upgradeVersion.Version, version.Version) {
		return false
	}
	if !utils.VersionBetween(upgradeVersion.MinimumRancherVersion, rancherVersion, upgradeVersion.MaximumRancherVersion) {
		return false
	}
	if upgradeVersion.UpgradeFrom != "" {
		satisfiesRange, err := utils.VersionSatisfiesRange(version.Version, upgradeVersion.UpgradeFrom)
		if err != nil {
			return false
		}
		return satisfiesRange
	}
	return true
}
