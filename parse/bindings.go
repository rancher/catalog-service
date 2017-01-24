package parse

import (
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/utils"
	"github.com/docker/libcompose/yaml"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/rancher-compose/preprocess"
)

func Bindings(contents []byte) (map[string]model.Bindings, error) {
	config, err := config.CreateConfig(contents)
	if err != nil {
		return nil, err
	}

	rawServiceMap := config.Services

	preProcessServiceMap := preprocess.PreprocessServiceMap(nil)
	rawServiceMap, err = preProcessServiceMap(rawServiceMap)
	if err != nil {
		return nil, err
	}

	bindingsMap := map[string]model.Bindings{}
	for serviceName, service := range rawServiceMap {
		var bindings model.Bindings

		var labels yaml.MaporEqualSlice
		if rawLabels, ok := service["labels"]; ok {
			if err = utils.Convert(rawLabels, &labels); err != nil {
				return nil, err
			}
			bindings.Labels = labels.ToMap()
		}
		if rawPorts, ok := service["ports"]; ok {
			if err = utils.Convert(rawPorts, &bindings.Ports); err != nil {
				return nil, err
			}
		}
		bindingsMap[serviceName] = bindings
	}

	return bindingsMap, nil
}
