package manager

import (
	"path"
	"strings"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/yaml.v2"
)

func traverseFiles(files *object.FileIter) ([]model.Template, []model.Version, error) {
	templateIndex := map[string]*model.Template{}
	versionsIndex := map[string]*model.Version{}
	if err := files.ForEach(func(f *object.File) error {
		templatesBase, parsedCorrectly := getTemplatesBase(f.Name)
		if !parsedCorrectly {
			return nil
		}

		dir, filename := path.Split(f.Name)

		switch {
		case filename == "config.yml":
			_, templateFolderName, parsedCorrectly := parse.TemplatePath(f.Name)
			if !parsedCorrectly {
				return nil
			}
			contents, err := f.Contents()
			if err != nil {
				return err
			}
			//var templateConfig TemplateConfig
			var template model.Template
			if err = yaml.Unmarshal([]byte(contents), &template); err != nil {
				return err
			}
			template.Base = templatesBase
			template.FolderName = templateFolderName
			if existingTemplate, ok := templateIndex[dir]; ok {
				template.Icon = existingTemplate.Icon
			}
			templateIndex[dir] = &template
			// TODO: just move this to the end of the function
			//templates = append(templates, template)
		case strings.HasPrefix(filename, "catalogIcon"):
			_, _, parsedCorrectly := parse.TemplatePath(f.Name)
			if !parsedCorrectly {
				return nil
			}
			contents, err := f.Contents()
			if err != nil {
				return err
			}
			if _, ok := templateIndex[dir]; !ok {
				templateIndex[dir] = &model.Template{}
			}
			templateIndex[dir].Icon = []byte(contents)
			//case strings.ToLower(filename):
			// TODO: determine if this is in template or version
		default:
			_, templateFolderName, revision, parsedCorrectly := parse.VersionPath(f.Name)
			if !parsedCorrectly {
				return nil
			}
			contents, err := f.Contents()
			if err != nil {
				return err
			}
			if _, ok := versionsIndex[dir]; !ok {
				versionsIndex[dir] = &model.Version{}
				versionsIndex[dir].Template = templateFolderName
				versionsIndex[dir].Revision = revision
			}
			versionsIndex[dir].Files = append(versionsIndex[dir].Files, model.File{
				Name:     filename,
				Contents: contents,
			})

		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	templates := []model.Template{}
	for _, template := range templateIndex {
		templates = append(templates, *template)
	}

	versions := []model.Version{}
	for _, version := range versionsIndex {
		var rancherCompose string
		for _, file := range version.Files {
			if file.Name == "rancher-compose.yml" {
				rancherCompose = file.Contents
			}
		}
		if rancherCompose == "" {
			continue
		}

		catalogInfo, err := parse.CatalogInfoFromRancherCompose([]byte(rancherCompose))
		if err != nil {
			return nil, nil, err
		}
		catalogInfo.Template = version.Template
		catalogInfo.Revision = version.Revision
		catalogInfo.Files = version.Files
		versions = append(versions, catalogInfo)
	}

	return templates, versions, nil
}
