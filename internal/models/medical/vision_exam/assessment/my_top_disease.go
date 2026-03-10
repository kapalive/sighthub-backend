package assessment

type MyTopDisease struct {
	IDMyTopDisease int64   `gorm:"column:id_my_top_disease;primaryKey;autoIncrement" json:"id_my_top_disease"`
	LevelID        int64   `gorm:"column:level_id;not null" json:"level_id"`
	Type           string  `gorm:"column:type;size:50;not null" json:"type"`
	Code           string  `gorm:"column:code;size:20;not null" json:"code"`
	Title          string  `gorm:"column:title;size:255;not null" json:"title"`
	GroupSet       *string `gorm:"column:group_set;size:255" json:"group_set"`
}

func (MyTopDisease) TableName() string { return "my_top_disease" }
