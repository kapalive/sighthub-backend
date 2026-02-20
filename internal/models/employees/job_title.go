package employees

import "fmt"

type JobTitle struct {
	IDJobTitle int     `gorm:"column:id_job_title;primaryKey"                     json:"id_job_title"`
	Title      string  `gorm:"column:title;type:varchar(100);not null"            json:"title"`
	Doctor     bool    `gorm:"column:doctor;not null;default:false"               json:"doctor"`
	ShortTitle *string `gorm:"column:short_title;type:varchar(3)"                 json:"short_title,omitempty"`
}

func (JobTitle) TableName() string { return "job_title" }

func (j *JobTitle) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_job_title": j.IDJobTitle,
		"title":        j.Title,
		"doctor":       j.Doctor,
		"short_title":  j.ShortTitle,
	}
}

func (j *JobTitle) String() string {
	short := ""
	if j.ShortTitle != nil {
		short = *j.ShortTitle
	}
	return fmt.Sprintf("<JobTitle %s | Doctor: %t | Short: %s>", j.Title, j.Doctor, short)
}
