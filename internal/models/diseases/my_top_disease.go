// internal/models/diseases/my_top_disease.go
package diseases

// MyTopDisease ↔ table: my_top_disease (user's saved favourite diagnoses)
type MyTopDisease struct {
	IDMyTopDisease int64   `gorm:"column:id_my_top_disease;primaryKey;autoIncrement" json:"id_my_top_disease"`
	LevelID        int64   `gorm:"column:level_id;not null"                          json:"level_id"`
	Type           string  `gorm:"column:type;type:varchar(50);not null"             json:"type"`
	Code           string  `gorm:"column:code;type:varchar(20);not null"             json:"code"`
	Title          string  `gorm:"column:title;type:varchar(255);not null"           json:"title"`
	GroupSet       *string `gorm:"column:group_set;type:varchar(255)"                json:"group_set,omitempty"`
}

func (MyTopDisease) TableName() string { return "my_top_disease" }

func (m *MyTopDisease) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_my_top_disease": m.IDMyTopDisease,
		"level_id":          m.LevelID,
		"type":              m.Type,
		"code":              m.Code,
		"title":             m.Title,
		"group_set":         m.GroupSet,
	}
}
