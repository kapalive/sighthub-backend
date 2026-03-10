// internal/repository/medical_repo/diseases_repo/chapter_disease.go
package diseases_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/diseases"
)

type ChapterDiseaseRepo struct{ DB *gorm.DB }

func NewChapterDiseaseRepo(db *gorm.DB) *ChapterDiseaseRepo {
	return &ChapterDiseaseRepo{DB: db}
}

func (r *ChapterDiseaseRepo) GetAll() ([]diseases.ChapterDisease, error) {
	var list []diseases.ChapterDisease
	if err := r.DB.Order("letter").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ChapterDiseaseRepo) GetByID(id int64) (*diseases.ChapterDisease, error) {
	var c diseases.ChapterDisease
	if err := r.DB.First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *ChapterDiseaseRepo) Search(q string) ([]diseases.ChapterDisease, error) {
	var list []diseases.ChapterDisease
	if err := r.DB.Where("title ILIKE ?", "%"+q+"%").
		Order("letter").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
