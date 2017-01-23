package manager

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rancher/catalog-service/model"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
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
		versionModel := model.VersionModel{
			Version: version,
		}
		if err = db.Create(&versionModel).Error; err != nil {
			return err
		}
		for _, file := range version.Files {
			file.VersionID = versionModel.ID
			if err = db.Create(&model.FileModel{
				File: file,
			}).Error; err != nil {
				return err
			}
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
