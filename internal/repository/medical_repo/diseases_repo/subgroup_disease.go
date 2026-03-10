// internal/repository/medical_repo/diseases_repo/subgroup_disease.go
package diseases_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/diseases"
)

type SubgroupDiseaseRepo struct{ DB *gorm.DB }

func NewSubgroupDiseaseRepo(db *gorm.DB) *SubgroupDiseaseRepo {
	return &SubgroupDiseaseRepo{DB: db}
}

func (r *SubgroupDiseaseRepo) GetByGroupID(groupID int64) ([]diseases.SubgroupDisease, error) {
	var list []diseases.SubgroupDisease
	if err := r.DB.Where("group_disease_id_group_disease = ?", groupID).
		Order("code").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *SubgroupDiseaseRepo) GetByID(id int64) (*diseases.SubgroupDisease, error) {
	var s diseases.SubgroupDisease
	if err := r.DB.First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SubgroupDiseaseRepo) Search(q string) ([]diseases.SubgroupDisease, error) {
	var list []diseases.SubgroupDisease
	if err := r.DB.Where("title ILIKE ? OR code ILIKE ?", "%"+q+"%", "%"+q+"%").
		Order("code").Limit(50).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
