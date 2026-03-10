// internal/repository/medical_repo/diseases_repo/group_disease.go
package diseases_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/diseases"
)

type GroupDiseaseRepo struct{ DB *gorm.DB }

func NewGroupDiseaseRepo(db *gorm.DB) *GroupDiseaseRepo {
	return &GroupDiseaseRepo{DB: db}
}

func (r *GroupDiseaseRepo) GetByChapterID(chapterID int64) ([]diseases.GroupDisease, error) {
	var list []diseases.GroupDisease
	if err := r.DB.Where("chapter_disease_id_chapter_disease = ?", chapterID).
		Order("code").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GroupDiseaseRepo) GetByID(id int64) (*diseases.GroupDisease, error) {
	var g diseases.GroupDisease
	if err := r.DB.First(&g, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

func (r *GroupDiseaseRepo) Search(q string) ([]diseases.GroupDisease, error) {
	var list []diseases.GroupDisease
	if err := r.DB.Where("title ILIKE ? OR code ILIKE ?", "%"+q+"%", "%"+q+"%").
		Order("code").Limit(50).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
