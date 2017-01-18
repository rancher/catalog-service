package model

import "github.com/rancher/go-rancher/client"

type CatalogError struct {
	client.Resource
	Status  string `json:"status"`
	Message string `json:"message"`
}
