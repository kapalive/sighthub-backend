package marketing

type VisitReason struct {
	IDVisitReasons int     `gorm:"column:id_visit_reasons;primaryKey" json:"id_visit_reasons"`
	Title          string  `gorm:"column:title;not null"              json:"title"`
	Description    *string `gorm:"column:description;type:text"       json:"description,omitempty"`
}

func (VisitReason) TableName() string { return "visit_reasons" }
