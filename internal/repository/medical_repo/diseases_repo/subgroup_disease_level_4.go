// internal/repository/medical_repo/diseases_repo/subgroup_disease_level_4.go
package diseases_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/diseases"
)

type SubgroupDiseaseLevel4Repo struct{ DB *gorm.DB }

func NewSubgroupDiseaseLevel4Repo(db *gorm.DB) *SubgroupDiseaseLevel4Repo {
	return &SubgroupDiseaseLevel4Repo{DB: db}
}

func (r *SubgroupDiseaseLevel4Repo) GetBySubgroupID(subgroupID int64) ([]diseases.SubgroupDiseaseLevel4, error) {
	var list []diseases.SubgroupDiseaseLevel4
	if err := r.DB.Where("subgroup_disease_id_subgroup_disease = ?", subgroupID).
		Order("code").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *SubgroupDiseaseLevel4Repo) GetByID(id int64) (*diseases.SubgroupDiseaseLevel4, error) {
	var s diseases.SubgroupDiseaseLevel4
	if err := r.DB.First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SubgroupDiseaseLevel4Repo) GetByCode(code string) (*diseases.SubgroupDiseaseLevel4, error) {
	var s diseases.SubgroupDiseaseLevel4
	if err := r.DB.Where("code = ?", code).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}
