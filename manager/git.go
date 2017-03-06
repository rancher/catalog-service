package manager

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/git"
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

func (m *Manager) prepareRepoPath(catalog model.Catalog, update bool) (string, string, error) {
	branch := catalog.Branch
	if catalog.Branch == "" {
		branch = "master"
	}

	sum := md5.Sum([]byte(catalog.URL + branch))
	repoBranchHash := hex.EncodeToString(sum[:])
	repoPath := path.Join(m.cacheRoot, catalog.EnvironmentId, repoBranchHash)

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return "", "", err
	}

	empty, err := dirEmpty(repoPath)
	if err != nil {
		return "", "", err
	}

	if empty {
		if err = git.Clone(repoPath, catalog.URL, branch); err != nil {
			return "", "", err
		}
	} else {
		if update {
			if err = git.Update(repoPath, branch); err != nil {
				// Ignore error unless running in strict mode
				if m.strict {
					return "", "", err
				}
				log.Errorf("Failed to update existing repo cache: %v", err)
			}
		}
	}

	commit, err := git.HeadCommit(repoPath)
	return repoPath, commit, err
}
