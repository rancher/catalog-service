package manager

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/rancher/catalog-service/model"
)

func (m *Manager) prepareRepoPath(catalog model.Catalog) (string, error) {
	branch := catalog.Branch
	if catalog.Branch == "" {
		branch = "master"
	}

	sum := md5.Sum([]byte(catalog.URL + branch))
	repoBranchHash := hex.EncodeToString(sum[:])
	repoPath := path.Join(m.cacheRoot, catalog.EnvironmentId, repoBranchHash)

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return "", err
	}

	f, err := os.Open(repoPath)
	if err != nil {
		return "", err
	}

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		cmd := exec.Command("git", "clone", "-b", branch, "--single-branch", catalog.URL, repoPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return "", err
		}
	}

	f.Close()

	cmd := exec.Command("git", "-C", repoPath, "fetch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return "", err
	}

	cmd = exec.Command("git", "-C", repoPath, "checkout", fmt.Sprintf("origin/%s", branch))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return "", err
	}

	return repoPath, nil
}
