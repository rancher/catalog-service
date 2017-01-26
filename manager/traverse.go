package manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"gopkg.in/yaml.v2"
)

func traverseFiles(repoPath string) ([]model.Template, []model.Version, error) {
	templateIndex := map[string]*model.Template{}
	versionsIndex := map[string]*model.Version{}

	if err := filepath.Walk(repoPath, func(fullPath string, f os.FileInfo, err error) error {
		relativePath, err := filepath.Rel(repoPath, fullPath)
		if err != nil {
			return err
		}

		templatesBase, parsedCorrectly := getTemplatesBase(relativePath)
		if !parsedCorrectly {
			return nil
		}

		dir, filename := path.Split(relativePath)

		switch {
		case filename == "config.yml":
			_, templateFolderName, parsedCorrectly := parse.TemplatePath(relativePath)
			if !parsedCorrectly {
				return nil
			}
			contents, err := ioutil.ReadFile(fullPath)
			if err != nil {
				return nil
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
				template.IconFilename = existingTemplate.IconFilename
			}
			templateIndex[dir] = &template
			// TODO: just move this to the end of the function
			//templates = append(templates, template)
		case strings.HasPrefix(filename, "catalogIcon"):
			_, _, parsedCorrectly := parse.TemplatePath(relativePath)
			if !parsedCorrectly {
				return nil
			}
			contents, err := ioutil.ReadFile(fullPath)
			if err != nil {
				return nil
				return err
			}
			if _, ok := templateIndex[dir]; !ok {
				templateIndex[dir] = &model.Template{}
			}
			templateIndex[dir].Icon = []byte(contents)
			templateIndex[dir].IconFilename = filename
			//case strings.ToLower(filename):
			// TODO: determine if README is in template or version
		default:
			_, templateFolderName, revision, parsedCorrectly := parse.VersionPath(relativePath)
			if !parsedCorrectly {
				return nil
			}

			fmt.Println(templateFolderName, revision)
			contents, err := ioutil.ReadFile(fullPath)
			if err != nil {
				return nil
				return err
			}
			if _, ok := versionsIndex[dir]; !ok {
				versionsIndex[dir] = &model.Version{}
				versionsIndex[dir].Template = templateFolderName
				versionsIndex[dir].Revision = revision
			}
			versionsIndex[dir].Files = append(versionsIndex[dir].Files, model.File{
				Name:     filename,
				Contents: string(contents),
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
		var readme string
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
		newVersion := *version
		if rancherCompose != "" {
			var err error
			newVersion, err = parse.CatalogInfoFromRancherCompose([]byte(rancherCompose))
			if err != nil {
				return nil, nil, err
			}
			newVersion.Template = version.Template
			newVersion.Revision = version.Revision
			newVersion.Files = version.Files
		}
		newVersion.Readme = readme
		versions = append(versions, newVersion)
	}

	return templates, versions, nil
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
