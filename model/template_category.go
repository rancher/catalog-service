package model

type TemplateCategory struct {
	TemplateId uint `sql:"type:integer REFERENCES catalog_template(id) ON DELETE CASCADE"`
	CategoryId uint `sql:"type:integer REFERENCES catalog_category(id) ON DELETE CASCADE"`
}

type TemplateCategoryModel struct {
	Base
	TemplateCategory
}
