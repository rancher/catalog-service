package parse

import (
	"strconv"
	"strings"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/utils"
	"github.com/rancher/catalog-service/model"
	"gopkg.in/yaml.v2"
)

func CatalogInfoFromRancherCompose(contents []byte) (model.Version, error) {
	cfg, err := config.CreateConfig(contents)
	if err != nil {
		return model.Version{}, err
	}
	var rawCatalogConfig interface{}

	if cfg.Version == "2" && cfg.Services[".catalog"] != nil {
		rawCatalogConfig = cfg.Services[".catalog"]
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(contents, &data); err != nil {
		return model.Version{}, err
	}

	if data["catalog"] != nil {
		rawCatalogConfig = data["catalog"]
	} else if data[".catalog"] != nil {
		rawCatalogConfig = data[".catalog"]
	}

	if rawCatalogConfig != nil {
		var template model.Version
		if err := utils.Convert(rawCatalogConfig, &template); err != nil {
			return model.Version{}, err
		}
		return template, nil
	}

	return model.Version{}, nil
}

/*func TemplateURLPath(path string) (string, int, string, error) {
	split := strings.Split(path, "/")
	if len(split) < 3 {
		return "", "", 0, false
	}

	catalog := split[0]
	template := split[1]
	revision, err := strconv.Atoi(split[2])
	if err != nil {
		return "", "", 0, false
	}

	return catalog, template, revision, true

}*/

func ConfigPath(path string) (string, string, bool) {
	split := strings.Split(path, "/")
	if len(split) < 2 {
		return "", "", false
	}
	return split[0], split[1], true
}

func DiskPath(path string) (string, string, int, bool) {
	split := strings.Split(path, "/")
	if len(split) < 3 {
		return "", "", 0, false
	}

	catalog := split[0]
	template := split[1]
	revision, err := strconv.Atoi(split[2])
	if err != nil {
		return "", "", 0, false
	}

	return catalog, template, revision, true
}
