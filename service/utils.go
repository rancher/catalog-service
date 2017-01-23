package service

import (
	"fmt"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/model"
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
	if template.Base == "" {
		return fmt.Sprintf("%s:%s:%d", template.Catalog, template.FolderName, version.Revision)
	}
	return fmt.Sprintf("%s:%s*%s:%d", template.Catalog, template.Base, template.FolderName, version.Revision)
}

func templateId(template model.Template) string {
	// TODO: use logrus
	if template.FolderName == "" {
		fmt.Println("Missing FolderName")
	}
	if template.Base == "" {
		return fmt.Sprintf("%s:%s", template.Catalog, template.FolderName)
	}
	return fmt.Sprintf("%s:%s*%s", template.Catalog, template.Base, template.FolderName)
}

func addTemplateFieldsToVersion(version *model.Version, template *model.Template) *model.Version {
	version.Category = template.Category
	version.IsSystem = template.IsSystem
	version.Description = template.Description
	version.DefaultVersion = template.DefaultVersion
	//version.IconLink = template.IconLink
	version.Path = template.Path
	// TODO: finish
	return version
}

func templateResource(apiContext *api.ApiContext, template model.Template, versions []model.Version) *model.TemplateResource {
	templateId := templateId(template)

	versionLinks := map[string]string{}
	for _, version := range versions {
		route := versionId(template, version)
		link := apiContext.UrlBuilder.ReferenceByIdLink("template", route)
		versionLinks[version.Version] = URLEncoded(link)
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
		IconLink:     links["icon"],
		VersionLinks: versionLinks,
	}
}

func versionResource(apiContext *api.ApiContext, template model.Template, version model.Version, files []model.File) (*model.TemplateVersionResource, error) {
	templateId := templateId(template)
	versionId := versionId(template, version)

	filesMap := map[string]string{}
	for _, file := range files {
		filesMap[file.Name] = file.Contents
	}

	var bindings map[string]model.Bindings
	/*dockerCompose, ok := filesMap["docker-compose.yml"]
	if ok {
		var err error
		bindings, err = parse.Bindings([]byte(dockerCompose))
		if err != nil {
			return nil, err
		}
	}*/

	links := map[string]string{}
	links["icon"] = URLEncoded(apiContext.UrlBuilder.ReferenceByIdLink("template", fmt.Sprintf("%s?image", templateId)))

	return &model.TemplateVersionResource{
		Resource: client.Resource{
			Id:    versionId,
			Type:  "templateVersion",
			Links: links,
		},
		Version:  *addTemplateFieldsToVersion(&version, &template),
		Bindings: bindings,
		Files:    filesMap,
	}, nil
}
