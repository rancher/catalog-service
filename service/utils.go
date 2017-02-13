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

func generateVersionId(catalogName string, template model.Template, version model.Version) string {
	if template.Base == "" {
		return fmt.Sprintf("%s:%s:%d", catalogName, template.FolderName, version.Revision)
	}
	return fmt.Sprintf("%s:%s*%s:%d", catalogName, template.Base, template.FolderName, version.Revision)
}

func generateTemplateId(catalogName string, template model.Template) string {
	if template.Base == "" {
		return fmt.Sprintf("%s:%s", catalogName, template.FolderName)
	}
	return fmt.Sprintf("%s:%s*%s", catalogName, template.Base, template.FolderName)
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

func templateResource(apiContext *api.ApiContext, catalogName string, template model.Template, rancherVersion string) *model.TemplateResource {
	templateId := generateTemplateId(catalogName, template)

	versionLinks := map[string]string{}
	for _, version := range template.Versions {
		if utils.VersionBetween(version.MinimumRancherVersion, rancherVersion, version.MaximumRancherVersion) {
			route := generateVersionId(catalogName, template, version)
			link := apiContext.UrlBuilder.ReferenceByIdLink("template", route)
			versionLinks[version.Version] = URLEncoded(link)
		}
	}

	links := map[string]string{}
	links["icon"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?image", templateId)))
	if template.Readme != "" {
		links["readme"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?readme", templateId)))
	}

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

func versionResource(apiContext *api.ApiContext, catalogName string, template model.Template, version model.Version, rancherVersion string) (*model.TemplateVersionResource, error) {
	templateId := generateTemplateId(catalogName, template)
	versionId := generateVersionId(catalogName, template, version)

	filesMap := map[string]string{}
	for _, file := range version.Files {
		filesMap[file.Name] = file.Contents
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
	if version.Readme != "" {
		links["readme"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?readme", versionId)))
	}

	upgradeVersionLinks := map[string]string{}
	for _, upgradeVersion := range template.Versions {
		if showUpgradeVersion(version, upgradeVersion, rancherVersion) {
			route := generateVersionId(catalogName, template, upgradeVersion)
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
