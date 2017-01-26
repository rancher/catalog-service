package manager

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
)

func (m *Manager) prepareRepoPath(name string, config CatalogConfig) (string, error) {
	branch := config.Branch
	if config.Branch == "" {
		branch = "master"
	}

	sum := md5.Sum([]byte(config.URL + branch))
	repoBranchHash := hex.EncodeToString(sum[:])
	repoPath := path.Join(m.cacheRoot, config.EnvironmentId, repoBranchHash)

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return "", err
	}

	f, err := os.Open(repoPath)
	if err != nil {
		return "", err
	}

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		cmd := exec.Command("git", "clone", "-b", branch, "--single-branch", config.URL, repoPath)
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
