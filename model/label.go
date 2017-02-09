package model

type Label struct {
	TemplateId uint `sql:"type:integer REFERENCES catalog_template(id) ON DELETE CASCADE"`

	Key   string
	Value string
}

type LabelModel struct {
	Base
	Label
}
