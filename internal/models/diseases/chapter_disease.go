// internal/models/diseases/chapter_disease.go
package diseases

// ChapterDisease ↔ table: chapter_disease (ICD-10 top level)
type ChapterDisease struct {
	IDChapterDisease int64  `gorm:"column:id_chapter_disease;primaryKey;autoIncrement" json:"id_chapter_disease"`
	Letter           string `gorm:"column:letter;type:char(1);not null"                json:"letter"`
	Title            string `gorm:"column:title;type:varchar(255);not null"            json:"title"`
}

func (ChapterDisease) TableName() string { return "chapter_disease" }

func (c *ChapterDisease) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_chapter_disease": c.IDChapterDisease,
		"letter":             c.Letter,
		"title":              c.Title,
	}
}
