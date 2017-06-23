package git

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func Clone(path, url, branch string) error {
	return runcmd("git", "clone", "-b", branch, "--single-branch", url, path)
}

func Update(path, branch string) error {
	if err := runcmd("git", "-C", path, "fetch"); err != nil {
		return err
	}
	return runcmd("git", "-C", path, "checkout", fmt.Sprintf("origin/%s", branch))
}

func HeadCommit(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "HEAD")
	output, err := cmd.Output()
	return strings.Trim(string(output), "\n"), err
}

func formatGitURL(url, branch string) string {
	splitURL := strings.Split(url, "/")

	if strings.HasPrefix(url, "https://github.com") && len(splitURL) > 4 {
		org := splitURL[3]
		repo := strings.TrimSuffix(splitURL[4], ".git")

		return fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", org, repo, branch)
	} else if strings.HasPrefix(url, "https://git.rancher.io") && len(splitURL) > 3 {
		repo := strings.TrimSuffix(splitURL[3], ".git")

		return fmt.Sprintf("https://git.rancher.io/repos/%s/commits/%s", repo, branch)
	}

	return ""
}

func RemoteShaChanged(url, branch, sha string) bool {
	formattedURL := formatGitURL(url, branch)

	if formattedURL == "" {
		return true
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.Header.Set("Accept", "application/vnd.github.chitauri-preview+sha")
			req.Header.Set("If-None-Match", fmt.Sprintf("\"%s\"", sha))

			return nil
		},
	}
	req, _ := http.NewRequest("GET", formattedURL, nil)
	req.Header.Set("Accept", "application/vnd.github.chitauri-preview+sha")
	req.Header.Set("If-None-Match", fmt.Sprintf("\"%s\"", sha))
	res, _ := client.Do(req)

	if res.StatusCode == 304 {
		return false
	}

	return true
}

func IsValid(url string) bool {
	err := runcmd("git", "ls-remote", url)
	return (err == nil)
}

func runcmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if log.GetLevel() >= log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}
