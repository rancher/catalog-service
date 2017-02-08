package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Clone(path, url, branch string) error {
	cmd := exec.Command("git", "clone", "-b", branch, "--single-branch", url, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Update(path, branch string) error {
	cmd := exec.Command("git", "-C", path, "fetch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "-C", path, "checkout", fmt.Sprintf("origin/%s", branch))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func HeadCommit(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "HEAD")
	output, err := cmd.Output()
	return strings.Trim(string(output), "\n"), err
}
