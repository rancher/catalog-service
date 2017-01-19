package service

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/rancher/catalog-service/manager"
	"github.com/rancher/catalog-service/model"
	"github.com/rancher/go-rancher/api"
	"github.com/rancher/go-rancher/client"
)

// MuxWrapper is a wrapper over the mux router that returns 503 until catalog is ready
type MuxWrapper struct {
	IsReady bool
	Router  *mux.Router
}

func (httpWrapper *MuxWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpWrapper.Router.ServeHTTP(w, r)
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var routes = []Route{
	{
		"GetTemplates",
		"GET",
		"/v1-catalog/templates",
		getTemplates,
	},
	{
		"GetTemplate",
		"GET",
		"/v1-catalog/templates/{catalog_template_version}",
		getTemplate,
	},
	{
		"GetCatalogs",
		"GET",
		"/v1-catalog/catalogs",
		getCatalogs,
	},
	{
		"GetCatalog",
		"GET",
		"/v1-catalog/catalogs/{catalog}",
		getCatalog,
	},
	{
		"GetCatalogTemplates",
		"GET",
		"/v1-catalog/catalogs/{catalog}/templates",
		getCatalogTemplates,
	},
	{
		"RefreshCatalog",
		"POST",
		"/v1-catalog/templates",
		refreshCatalog,
	},
}

// TODO
var schemas *client.Schemas

// TODO
var m *manager.Manager
var db *gorm.DB

func NewRouter(manager *manager.Manager, gormDb *gorm.DB) *mux.Router {
	// TODO
	m = manager
	db = gormDb

	schemas := &client.Schemas{}

	apiVersion := schemas.AddType("apiVersion", client.Resource{})
	apiVersion.CollectionMethods = []string{}

	schemas.AddType("schema", client.Schema{})

	schemas.AddType("catalog", model.CatalogResource{})

	template := schemas.AddType("template", model.TemplateResource{})
	template.CollectionActions = map[string]client.Action{
		"refresh": {},
	}
	delete(template.ResourceFields, "icon")

	templateVersion := schemas.AddType("templateVersion", model.TemplateVersionResource{})
	templateVersion.CollectionMethods = []string{}
	// TODO: move to generic files map
	delete(template.ResourceFields, "dockerCompose")
	delete(template.ResourceFields, "rancherCompose")

	err := schemas.AddType("error", model.CatalogError{})
	err.CollectionMethods = []string{}

	// API framework routes
	router := mux.NewRouter().StrictSlash(true)

	router.Methods("GET").Path("/").Handler(api.VersionsHandler(schemas, "v1-catalog"))
	router.Methods("GET").Path("/v1-catalog/schemas").Handler(api.SchemasHandler(schemas))
	router.Methods("GET").Path("/v1-catalog/schemas/{id}").Handler(api.SchemaHandler(schemas))
	router.Methods("GET").Path("/v1-catalog").Handler(api.VersionHandler(schemas, "v1-catalog"))

	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(api.ApiHandler(schemas, route.HandlerFunc))
	}

	router.GetRoute("RefreshCatalog").Queries("action", "refresh")

	return router
}
