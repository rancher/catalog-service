package manager

type CatalogType int

const (
	CatalogTypeRancher CatalogType = iota
	CatalogTypeHelmObjectRepo
	CatalogTypeHelmGitRepo
	CatalogTypeInvalid
)

type CatalogConfig struct {
	URL    string
	Branch string
	Kind   string
}
