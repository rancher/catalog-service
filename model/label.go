package model

import "github.com/jinzhu/gorm"

type Label struct {
	TemplateId uint `sql:"type:integer REFERENCES catalog_template(id) ON DELETE CASCADE"`

	Key   string
	Value string
}

type LabelModel struct {
	Base
	Label
}

func lookupLabels(db *gorm.DB, templateId uint) map[string]string {
	var labelModels []LabelModel
	db.Where(&LabelModel{
		Label: Label{
			TemplateId: templateId,
		},
	}).Find(&labelModels)

	labels := map[string]string{}
	for _, label := range labelModels {
		labels[label.Key] = label.Value
	}

	return labels
}
