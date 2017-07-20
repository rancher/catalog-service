package git

import (
	"fmt"
	"net/http"
	"net/url"
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

func formatGitURL(endpoint, branch string) string {
	formattedURL := ""
	if u, err := url.Parse(endpoint); err == nil {
		pathParts := strings.Split(u.Path, "/")
		switch strings.Split(u.Host, ":")[0] {
		case "github.com":
			if len(pathParts) >= 3 {
				org := pathParts[1]
				repo := strings.TrimSuffix(pathParts[2], ".git")
				formattedURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", org, repo, branch)
			}
		case "git.rancher.io":
			repo := strings.TrimSuffix(pathParts[1], ".git")
			u.Path = fmt.Sprintf("/repos/%s/commits/%s", repo, branch)
			formattedURL = u.String()
		}
	}
	return formattedURL
}

func RemoteShaChanged(url, branch, sha, uuid string) bool {
	formattedURL := formatGitURL(url, branch)

	if formattedURL == "" {
		return true
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.Header.Set("Accept", "application/vnd.github.chitauri-preview+sha")
			req.Header.Set("If-None-Match", fmt.Sprintf("\"%s\"", sha))
			if uuid != "" {
				req.Header.Set("X-Install-Uuid", uuid)
			}
			return nil
		},
	}
	req, err := http.NewRequest("GET", formattedURL, nil)
	if err != nil {
		return true
	}
	req.Header.Set("Accept", "application/vnd.github.chitauri-preview+sha")
	req.Header.Set("If-None-Match", fmt.Sprintf("\"%s\"", sha))
	if uuid != "" {
		req.Header.Set("X-Install-Uuid", uuid)
	}
	res, err := client.Do(req)
	if err != nil {
		return true
	}
	defer res.Body.Close()

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
