// internal/models/general/group_pqrs.go
package general

import "fmt"

type GroupPQRS struct {
	IDGroupPQRS int    `gorm:"column:id_group_pqrs;primaryKey;autoIncrement" json:"id_group_pqrs"`
	Title       string `gorm:"column:title;type:varchar(255);not null"       json:"title"`
}

func (GroupPQRS) TableName() string { return "group_pqrs" }

func (g *GroupPQRS) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_group_pqrs": g.IDGroupPQRS,
		"title":         g.Title,
	}
}

func (g *GroupPQRS) String() string { return fmt.Sprintf("<GroupPQRS %s>", g.Title) }
