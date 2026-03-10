package patients

// PreferredLanguage ⇄ table: preferred_language
type PreferredLanguage struct {
	IDPreferredLanguage int    `gorm:"column:id_preferred_language;primaryKey;autoIncrement" json:"id_preferred_language"`
	Language            string `gorm:"column:language;type:varchar(100);not null"            json:"language"`
}

func (PreferredLanguage) TableName() string { return "preferred_language" }
