package manager

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/catalog-service/parse"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/yaml.v2"
)

// TODO: move elsewhere
type CatalogConfig struct {
	URL    string
	Branch string
	// TODO: remove this
	EnvironmentId string
}

type Manager struct {
	cacheRoot string
	diskLocks map[string]*sync.Mutex
	//catalogURLs []string
	config map[string]CatalogConfig
	db     *gorm.DB
}

//func NewManager(cacheRoot string, catalogURLs []string) *Manager {
func NewManager(cacheRoot string, config map[string]CatalogConfig, db *gorm.DB) *Manager {
	return &Manager{
		cacheRoot: cacheRoot,
		config:    config,
		db:        db,
	}
}

func (m *Manager) RefreshAll() error {
	for catalogName, catalogConfig := range m.config {
		if err := m.refreshCatalog(catalogName, catalogConfig, m.db); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Refresh(environmentId string) error {
	for catalogName, catalogConfig := range m.config {
		if catalogConfig.EnvironmentId == environmentId {
			if err := m.refreshCatalog(catalogName, catalogConfig, m.db); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) refreshCatalog(name string, config CatalogConfig, db *gorm.DB) error {
	r, err := m.findOrClone(name, config, db)
	if err != nil {
		return err
	}

	ref, err := r.Head()
	if err != nil {
		return err
	}

	commit, err := r.Commit(ref.Hash())
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	templates, versions, err := traverseFiles(tree.Files())
	if err != nil {
		return err
	}

	for _, template := range templates {
		fmt.Println("#", template, template.Prefix)
	}

	for _, version := range versions {
		fmt.Println("@", version.Template, version.Revision)
	}

	// TODO: do not need to always create catalog
	if err = db.Create(&model.CatalogModel{
		Catalog: model.Catalog{
			Name:          name,
			URL:           config.URL,
			Branch:        config.Branch,
			EnvironmentId: config.EnvironmentId,
		},
	}).Error; err != nil {
		return err
	}

	for _, template := range templates {
		template.Catalog = name
		template.EnvironmentId = config.EnvironmentId
		if err = db.Create(&model.TemplateModel{
			Template: template,
		}).Error; err != nil {
			return err
		}
	}

	for _, version := range versions {
		version.Catalog = name
		version.EnvironmentId = config.EnvironmentId
		if err = db.Create(&model.VersionModel{
			Version: version,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) findOrClone(name string, config CatalogConfig, db *gorm.DB) (*git.Repository, error) {
	branch := config.Branch
	if config.Branch == "" {
		branch = "master"
	}

	sum := md5.Sum([]byte(config.URL + branch))
	repoBranchHash := hex.EncodeToString(sum[:])
	repoPath := path.Join(m.cacheRoot, config.EnvironmentId, repoBranchHash)

	if _, err := os.Stat(repoPath); err == nil {
		r, err := git.NewFilesystemRepository(repoPath)
		if err != nil {
			return nil, err
		}
		return r, r.Pull(&git.PullOptions{
			ReferenceName: plumbing.ReferenceName(config.Branch),
			SingleBranch:  true,
			Depth:         1,
		})
	}
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return nil, err
	}

	r, err := git.NewFilesystemRepository(repoPath)
	if err != nil {
		return nil, err
	}

	return r, r.Clone(&git.CloneOptions{
		URL:           config.URL,
		ReferenceName: plumbing.ReferenceName(config.Branch),
		SingleBranch:  true,
		Depth:         1,
	})
}

// TODO: move elsewhere
type TemplateConfig struct {
	Name           string            `yaml:"name"`
	Category       string            `yaml:"category"`
	Description    string            `yaml:"description"`
	Version        string            `yaml:"version"`
	Maintainer     string            `yaml:"maintainer"`
	License        string            `yaml:"license"`
	ProjectURL     string            `yaml:"projectURL"`
	IsSystem       string            `yaml:"isSystem"`
	DefaultVersion string            `yaml:"version"`
	Labels         map[string]string `yaml:"version"`
}

func getTemplatesPrefix(filename string) (string, bool) {
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

func traverseFiles(files *object.FileIter) ([]model.Template, []model.Version, error) {
	templates := []model.Template{}
	versions := []model.Version{}
	return templates, versions, files.ForEach(func(f *object.File) error {
		templatesPrefix, parsedCorrectly := getTemplatesPrefix(f.Name)
		if !parsedCorrectly {
			return nil
		}

		switch {
		case strings.HasSuffix(f.Name, "config.yml"):
			_, templateFolderName, parsedCorrectly := parse.ConfigPath(f.Name)
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
			template.Prefix = templatesPrefix
			template.FolderName = templateFolderName
			templates = append(templates, template)
		case strings.HasSuffix(f.Name, "docker-compose.yml"):
			// Save docker-compose.yml to version
			_, templateFolderName, revision, parsedCorrectly := parse.DiskPath(f.Name)
			if !parsedCorrectly {
				return nil
			}
			contents, err := f.Contents()
			if err != nil {
				return err
			}
			foundExisting := false
			for i, version := range versions {
				// TODO: need a better if
				if version.Template == templateFolderName && version.Revision == revision {
					versions[i].DockerCompose = contents
					foundExisting = true
				}
			}
			if !foundExisting {
				versions = append(versions, model.Version{
					Template:      templateFolderName,
					Revision:      revision,
					DockerCompose: contents,
				})
			}
		case strings.HasSuffix(f.Name, "rancher-compose.yml"):
			_, templateFolderName, revision, parsedCorrectly := parse.DiskPath(f.Name)
			if !parsedCorrectly {
				return nil
			}
			contents, err := f.Contents()
			if err != nil {
				return err
			}
			foundExisting := false
			for i, version := range versions {
				// TODO: need a better if
				if version.Template == templateFolderName && version.Revision == revision {
					foundExisting = true
					catalogInfo, err := parse.CatalogInfoFromRancherCompose([]byte(contents))
					if err != nil {
						return err
					}
					catalogInfo.Template = version.Template
					catalogInfo.Revision = version.Revision
					catalogInfo.DockerCompose = version.DockerCompose
					catalogInfo.RancherCompose = contents
					versions[i] = catalogInfo
				}
			}
			if !foundExisting {
				versions = append(versions, model.Version{
					Template:       templateFolderName,
					Revision:       revision,
					RancherCompose: contents,
				})
			}
			// Save rancher-compose.yml to version and parse catalog key
			/*_, templateFolderName, revision, parsedCorrectly := parse.DiskPath(f.Name)
			if !parsedCorrectly {
				return nil
			}
			contents, err := f.Contents()
			if err != nil {
				return err
			}
			template, err := parse.TemplateFromRancherCompose([]byte(contents))
			if err != nil {
				return err
			}
			_, templateFolderName, revision, parsedCorrectly := parse.DiskPath(f.Name)
			if parsedCorrectly {
				template.Revision = revision
				template.FolderName = templateFolderName
				templates = append(templates, template)
			}*/

		}

		return nil
	})
}
