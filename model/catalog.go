package model

import (
	"github.com/rancher/go-rancher/v2"
	catalogv1 "github.com/rancher/type/apis/catalog.cattle.io/v1"
)

const (
	ProjectLabel = "project.catalog.cattle.io"
)

type Catalog struct {
	EnvironmentId string `json:"environmentId"`

	Name   string `json:"name"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	Type   string `json:"type"`
	Kind   string `json:"kind"`
}

type CatalogModel struct {
	Base
	catalogv1.Catalog
}

type CatalogResource struct {
	client.Resource
	Catalog
}

type CatalogCollection struct {
	client.Collection
	Data []CatalogResource `json:"data,omitempty"`
}
