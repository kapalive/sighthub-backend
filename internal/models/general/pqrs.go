// internal/models/general/pqrs.go
package general

import "fmt"

type PQRS struct {
	IDPQRS       int64      `gorm:"column:id_pqrs;primaryKey;autoIncrement" json:"id_pqrs"`
	Code         string     `gorm:"column:code;type:varchar(10);not null"   json:"code"`
	Title        *string    `gorm:"column:title;type:varchar(255)"          json:"title,omitempty"`
	PQRSGroupID  *int       `gorm:"column:pqrs_group_id"                    json:"pqrs_group_id,omitempty"`
	GroupPQRSRef *GroupPQRS `gorm:"foreignKey:PQRSGroupID;references:IDGroupPQRS" json:"-"`
}

func (PQRS) TableName() string { return "pqrs" }

func (p *PQRS) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_pqrs": p.IDPQRS,
		"code":    p.Code,
	}
	if p.Title != nil {
		m["title"] = *p.Title
	} else {
		m["title"] = nil
	}
	if p.PQRSGroupID != nil {
		m["pqrs_group_id"] = *p.PQRSGroupID
	} else {
		m["pqrs_group_id"] = nil
	}
	return m
}

func (p *PQRS) String() string { return fmt.Sprintf("<PQRS %s>", p.Code) }
