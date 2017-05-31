package parse

import (
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/utils"
	"github.com/rancher/catalog-service/model"
	"gopkg.in/yaml.v2"
)

func CatalogInfoFromTemplateVersion(contents []byte) (model.Version, error) {

	var template model.Version
	if err := yaml.Unmarshal(contents, &template); err != nil {
		return model.Version{}, err
	}

	return template, nil
}

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
