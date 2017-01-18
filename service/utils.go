package service

import (
	"fmt"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"github.com/rancher/go-rancher/api"
	"github.com/rancher/go-rancher/client"
)

// TODO: this should return an error
func URLEncoded(str string) string {
	u, err := url.Parse(str)
	if err != nil {
		log.Errorf("Error encoding the url: %s , error: %v", str, err)
		return str
	}
	return u.String()
}

func versionId(template model.Template, version model.Version) string {
	// TODO: use logrus
	if template.FolderName == "" {
		fmt.Println("Missing FolderName")
	}
	return fmt.Sprintf("%s:%s:%d", template.Catalog, template.FolderName, version.Revision)
}

func templateId(template model.Template) string {
	// TODO: use logrus
	if template.FolderName == "" {
		fmt.Println("Missing FolderName")
	}
	return fmt.Sprintf("%s:%s", template.Catalog, template.FolderName)
}

func addTemplateFieldsToVersion(version *model.Version, template *model.Template) *model.Version {
	version.Category = template.Category
	version.IsSystem = template.IsSystem
	version.Description = template.Description
	version.DefaultVersion = template.DefaultVersion
	version.IconLink = template.IconLink
	version.Path = template.Path
	// TODO: finish
	return version
}

func templateResource(apiContext *api.ApiContext, template model.Template, versions []model.Version) *model.TemplateResource {
	versionLinks := map[string]string{}
	for _, version := range versions {
		route := versionId(template, version)
		link := apiContext.UrlBuilder.ReferenceByIdLink("template", route)
		versionLinks[version.Version] = URLEncoded(link)
	}
	return &model.TemplateResource{
		Resource: client.Resource{
			Id:   templateId(template),
			Type: "template",
		},
		Template:     template,
		VersionLinks: versionLinks,
	}
}

func versionResource(template model.Template, version model.Version) (*model.TemplateVersionResource, error) {
	bindings, err := parse.Bindings([]byte(version.DockerCompose))
	if err != nil {
		return nil, err
	}
	return &model.TemplateVersionResource{
		Resource: client.Resource{
			Id:   versionId(template, version),
			Type: "templateVersion",
		},
		Version:  *addTemplateFieldsToVersion(&version, &template),
		Bindings: bindings,
	}, nil
}
