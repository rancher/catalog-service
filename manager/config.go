package manager

import (
	"encoding/json"
	"io/ioutil"
)

type CatalogConfig struct {
	URL    string
	Branch string
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

	m.config = config["catalogs"]

	return nil
}
