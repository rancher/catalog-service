package manager

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path"

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

func (m *Manager) prepareRepoPath(catalog model.Catalog) (string, string, error) {
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
			return "", "", nil
		}
	} else {
		if err = git.Update(repoPath, branch); err != nil {
			return "", "", nil
		}
	}

	commit, err := git.HeadCommit(repoPath)
	return repoPath, commit, err
}
