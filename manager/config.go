package manager

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"
)

type CatalogType int

const (
	CatalogTypeRancher CatalogType = iota
	CatalogTypeHelmObjectRepo
	CatalogTypeHelmGitRepo
	CatalogTypeInvalid
)

type CatalogConfig struct {
	URL    string
	Branch string
	Kind   string
}

func (m *Manager) readConfig() error {
	configContents, err := ioutil.ReadFile(m.configFile)
	if err != nil {
		return err
	}

	var config map[string]map[string]CatalogConfig
	if err = json.Unmarshal(configContents, &config); err != nil {
		return err
	}

	for catalogName, catalogConfig := range config["catalogs"] {
		if u, err := url.Parse(catalogConfig.URL); err == nil {
			if strings.Split(u.Host, ":")[0] == "git.rancher.io" {
				u.Path = strings.Join([]string{strings.TrimRight(u.Path, "/"), m.uuid}, "/")
				catalogConfig.URL = u.String()
				config["catalogs"][catalogName] = catalogConfig
			}
		}
	}
	m.config = config["catalogs"]

	return nil
}
