package model

import (
	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/client"
)

//Catalog defines the properties of a template Catalog
/*type Catalog struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	CatalogLink string `json:"catalogLink"`
	URL         string `json:"uri"`
	State       string `json:"state"`
	LastUpdated string `json:"lastUpdated"`
	Message     string `json:"message"`
	URLBranch   string `json:"branch"`
}*/

type Catalog struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	Branch        string `json:"branch"`
	EnvironmentId string `json:"environmentId"`
}

type CatalogModel struct {
	gorm.Model
	Catalog
}

type CatalogResource struct {
	client.Resource
	Catalog
}

type CatalogCollection struct {
	client.Collection
	Data []CatalogResource `json:"data,omitempty"`
}
