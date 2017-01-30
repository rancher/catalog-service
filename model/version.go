package model

import (
	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/client"
)

// TODO: might need a Base field for filtering
// TODO: might need a FolderName field for filtering
type Version struct {
	Catalog               string `json:"catalogId"`
	EnvironmentId         string `json:"environmentId"`
	Template              string `json:"template"`
	Revision              int    `json:"revision"`
	Version               string `json:"version"`
	MinimumRancherVersion string `json:"minimumRancherVersion" yaml:"minimum_rancher_version"`
	MaximumRancherVersion string `json:"maximumRancherVersion" yaml:"maximum_rancher_version"`

	// TODO move to model
	Files     []File
	Questions []Question
	Readme    string `json:"readme"`
}

type VersionModel struct {
	gorm.Model
	Version
}

type TemplateVersionResource struct {
	client.Resource
	Version

	Bindings            map[string]Bindings `json:"bindings"`
	Files               map[string]string   `json:"files"`
	Questions           []Question          `json:"questions"`
	UpgradeVersionLinks map[string]string   `json:"upgradeVersionLinks"`
}
