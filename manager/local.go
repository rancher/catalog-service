package manager

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/git"
	"github.com/rancher/catalog-service/helm"
	"github.com/rancher/catalog-service/model"
)

func dirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func (m *Manager) prepareRepoPath(catalog model.Catalog, update bool) (string, string, CatalogType, error) {
	if catalog.Kind == "" || catalog.Kind == RancherTemplateType {
		return m.prepareGitRepoPath(catalog, update, CatalogTypeRancher)
	}
	if catalog.Kind == HelmTemplateType {
		if git.IsValid(catalog.URL) {
			return m.prepareGitRepoPath(catalog, update, CatalogTypeHelmGitRepo)
		}
		return m.prepareHelmRepoPath(catalog, update)
	}
	return "", "", CatalogTypeInvalid, fmt.Errorf("Unknown catalog kind=%s", catalog.Kind)
}

func (m *Manager) prepareHelmRepoPath(catalog model.Catalog, update bool) (string, string, CatalogType, error) {
	index, err := helm.DownloadIndex(catalog.URL)
	if err != nil {
		return "", "", CatalogTypeInvalid, err
	}

	repoPath := path.Join(m.cacheRoot, catalog.EnvironmentId, index.Hash)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return "", "", CatalogTypeInvalid, err
	}

	if err := helm.SaveIndex(index, repoPath); err != nil {
		return "", "", CatalogTypeInvalid, err
	}

	return repoPath, index.Hash, CatalogTypeHelmObjectRepo, nil
}

func (m *Manager) prepareGitRepoPath(catalog model.Catalog, update bool, catalogType CatalogType) (string, string, CatalogType, error) {
	branch := catalog.Branch
	if catalog.Branch == "" {
		branch = "master"
	}

	sum := md5.Sum([]byte(catalog.URL + branch))
	repoBranchHash := hex.EncodeToString(sum[:])
	repoPath := path.Join(m.cacheRoot, catalog.EnvironmentId, repoBranchHash)

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return "", "", catalogType, err
	}

	empty, err := dirEmpty(repoPath)
	if err != nil {
		return "", "", catalogType, err
	}

	if empty {
		if err = git.Clone(repoPath, catalog.URL, branch); err != nil {
			return "", "", catalogType, err
		}
	} else {
		if update {
			if git.RemoteShaChanged(catalog.URL, catalog.Branch, catalog.Commit) {
				if err = git.Update(repoPath, branch); err != nil {
					// Ignore error unless running in strict mode
					if m.strict {
						return "", "", catalogType, err
					}
					log.Errorf("Failed to update existing repo cache: %v", err)
				}
			}
		}
	}

	commit, err := git.HeadCommit(repoPath)
	return repoPath, commit, catalogType, err
}
