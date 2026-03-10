package marketing

type Form struct {
	IDForm    int    `gorm:"column:id_form;primaryKey;autoIncrement" json:"id_form"`
	Title     string `gorm:"column:title;size:255;not null"          json:"title"`
	PathFile  string `gorm:"column:path_file;type:text;not null"     json:"path_file"`
	IsActive  bool   `gorm:"column:is_active;default:true"           json:"is_active"`
}

func (Form) TableName() string { return "forms" }
