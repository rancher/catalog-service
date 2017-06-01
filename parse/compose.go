package parse

import (
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/utils"
	"github.com/rancher/catalog-service/model"
	"gopkg.in/yaml.v2"
)

func consolidateTemplateCamelSnakeCase(template model.Version) model.Version {

	if template.MinimumRancherVersionSnakeCase != "" {
		template.MinimumRancherVersion = template.MinimumRancherVersionSnakeCase
	} else if template.MinimumRancherVersionCamelCase != "" {
		template.MinimumRancherVersion = template.MinimumRancherVersionCamelCase
	}

	if template.MaximumRancherVersionSnakeCase != "" {
		template.MaximumRancherVersion = template.MaximumRancherVersionSnakeCase
	} else if template.MaximumRancherVersionCamelCase != "" {
		template.MaximumRancherVersion = template.MaximumRancherVersionCamelCase
	}

	if template.UpgradeFromSnakeCase != "" {
		template.UpgradeFrom = template.UpgradeFromSnakeCase
	} else if template.UpgradeFromCamelCase != "" {
		template.UpgradeFrom = template.UpgradeFromCamelCase
	}

	return template
}

func CatalogInfoFromTemplateVersion(contents []byte) (model.Version, error) {

	var template model.Version
	if err := yaml.Unmarshal(contents, &template); err != nil {
		return model.Version{}, err
	}

	return consolidateTemplateCamelSnakeCase(template), nil
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
		return consolidateTemplateCamelSnakeCase(template), nil
	}

	return model.Version{}, nil
}
