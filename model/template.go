package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rancher/go-rancher/client"
)

type Template struct {
	//CatalogID      string `json:"catalogId"`
	Name           string `json:"name"`
	Category       string `json:"category"`
	IsSystem       string `json:"isSystem"`
	Description    string `json:"description"`
	Version        string `json:"version"`
	DefaultVersion string `json:"defaultVersion"`
	IconLink       string `json:"iconLink"`
	//UpgradeVersionLinks map[string]string `json:"upgradeVersionLinks"`
	Path       string `json:"path"`
	Maintainer string `json:"maintainer"`
	License    string `json:"license"`
	ProjectURL string `json:"projectURL"`
	ReadmeLink string `json:"readmeLink"`
	//TemplateBase string `json:"templateBase"`
	//Labels                map[string]string      `json:"labels"`
	UpgradeFrom string `json:"upgradeFrom"`

	// TODO
	FolderName    string `json:"revision"`
	Catalog       string `json:"catalogId"`
	EnvironmentId string `json:"environmentId"`
	//Prefix        string `json:"prefix"`
	Base string `json:"templateBase"`
	Icon []byte `json:"icon"`
}

type TemplateModel struct {
	gorm.Model
	Template
}

type TemplateResource struct {
	client.Resource
	Template

	VersionLinks map[string]string `json:"versionLinks"`
}

type TemplateVersionResource struct {
	client.Resource
	Version

	Bindings map[string]Bindings `json:"bindings"`
	Files    map[string]string   `json:"files"`
}

type TemplateCollection struct {
	client.Collection
	Data []TemplateResource `json:"data,omitempty"`
}
