// internal/models/diseases/group_disease.go
package diseases

// GroupDisease ↔ table: group_disease (ICD-10 group)
type GroupDisease struct {
	IDGroupDisease                int64  `gorm:"column:id_group_disease;primaryKey;autoIncrement"          json:"id_group_disease"`
	ChapterDiseaseIDChapterDisease int64  `gorm:"column:chapter_disease_id_chapter_disease;not null"        json:"chapter_disease_id_chapter_disease"`
	Code                          string `gorm:"column:code;type:varchar(10);not null"                     json:"code"`
	Title                         string `gorm:"column:title;type:varchar(255);not null"                   json:"title"`

	Chapter *ChapterDisease `gorm:"foreignKey:ChapterDiseaseIDChapterDisease;references:IDChapterDisease" json:"-"`
}

func (GroupDisease) TableName() string { return "group_disease" }

func (g *GroupDisease) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_group_disease":                   g.IDGroupDisease,
		"chapter_disease_id_chapter_disease": g.ChapterDiseaseIDChapterDisease,
		"code":                               g.Code,
		"title":                              g.Title,
	}
}
