package manager

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"gopkg.in/yaml.v2"
)

func traverseFiles(repoPath string) ([]model.Template, error) {
	templateIndex := map[string]*model.Template{}

	if err := filepath.Walk(repoPath, func(fullPath string, f os.FileInfo, err error) error {
		relativePath, err := filepath.Rel(repoPath, fullPath)
		if err != nil {
			return err
		}

		templatesBase, parsedCorrectly := getTemplatesBase(relativePath)
		if !parsedCorrectly {
			return nil
		}

		_, filename := path.Split(relativePath)

		switch {
		case filename == "config.yml":
			base, templateName, parsedCorrectly := parse.TemplatePath(relativePath)
			if !parsedCorrectly {
				return nil
			}
			contents, err := ioutil.ReadFile(fullPath)
			if err != nil {
				// TODO
				return nil
				//return err
			}
			//var templateConfig TemplateConfig
			var template model.Template
			if err = yaml.Unmarshal([]byte(contents), &template); err != nil {
				return err
			}
			template.Base = templatesBase
			template.FolderName = templateName

			key := base + templateName

			if existingTemplate, ok := templateIndex[key]; ok {
				template.Icon = existingTemplate.Icon
				template.IconFilename = existingTemplate.IconFilename
				template.Versions = existingTemplate.Versions
			}
			templateIndex[key] = &template
			// TODO: just move this to the end of the function
			//templates = append(templates, template)
		case strings.HasPrefix(filename, "catalogIcon"):
			base, templateName, parsedCorrectly := parse.TemplatePath(relativePath)
			if !parsedCorrectly {
				return nil
			}

			contents, err := ioutil.ReadFile(fullPath)
			if err != nil {
				// TODO
				return nil
				//return err
			}

			key := base + templateName

			if _, ok := templateIndex[key]; !ok {
				templateIndex[key] = &model.Template{}
			}
			templateIndex[key].Icon = []byte(contents)
			templateIndex[key].IconFilename = filename
			//case strings.ToLower(filename):
			// TODO: determine if README is in template or version
		default:
			base, templateName, revision, parsedCorrectly := parse.VersionPath(relativePath)
			if !parsedCorrectly {
				return nil
			}

			contents, err := ioutil.ReadFile(fullPath)
			if err != nil {
				// TODO
				return nil
				//return err
			}

			key := base + templateName
			file := model.File{
				Name:     filename,
				Contents: string(contents),
			}

			if _, ok := templateIndex[key]; !ok {
				templateIndex[key] = &model.Template{}
			}
			for i, version := range templateIndex[key].Versions {
				if version.Revision == revision {
					templateIndex[key].Versions[i].Files = append(version.Files, file)
					return nil
				}
			}
			templateIndex[key].Versions = append(templateIndex[key].Versions, model.Version{
				Revision: revision,
				Files:    []model.File{file},
			})
		}

		return nil
	}); err != nil {
		return nil, err
	}

	templates := []model.Template{}
	for _, template := range templateIndex {
		for i, version := range template.Versions {
			var readme string
			//fmt.Println(template.FolderName, version.Revision, len(version.Files))
			for _, file := range version.Files {
				if strings.ToLower(file.Name) == "readme.md" {
					readme = file.Contents
				}
			}
			var rancherCompose string
			for _, file := range version.Files {
				if file.Name == "rancher-compose.yml" {
					rancherCompose = file.Contents
				}
			}
			newVersion := version
			if rancherCompose != "" {
				var err error
				newVersion, err = parse.CatalogInfoFromRancherCompose([]byte(rancherCompose))
				if err != nil {
					return nil, err
				}
				newVersion.Revision = version.Revision
				newVersion.Files = version.Files
			}
			newVersion.Readme = readme

			template.Versions[i] = newVersion
		}
		templates = append(templates, *template)
	}

	return templates, nil
}

func getTemplatesBase(filename string) (string, bool) {
	dir, _ := path.Split(filename)
	dirSplit := strings.Split(dir, "/")
	if len(dirSplit) < 2 {
		return "", false
	}
	firstDir := dirSplit[0]

	if firstDir == "templates" {
		return "", true
	}
	dirSplit = strings.Split(firstDir, "-")
	if len(dirSplit) != 2 {
		return "", false
	}
	return dirSplit[0], true
}
