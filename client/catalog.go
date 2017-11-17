package client

import (
	catalogv1 "github.com/rancher/type/apis/catalog.cattle.io/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	Config        string
	CatalogClient catalogv1.CatalogInterface
)

func NewCatalogClient(config string) (catalogv1.CatalogInterface, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", config)
	if err != nil {
		return nil, err
	}
	catalogInterface, err := catalogv1.NewForConfig(*cfg)
	if err != nil {
		return nil, err
	}
	return catalogInterface.Catalogs(""), nil
}
