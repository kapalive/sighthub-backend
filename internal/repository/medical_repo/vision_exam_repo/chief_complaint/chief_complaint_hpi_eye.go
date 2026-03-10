// internal/repository/medical_repo/vision_exam_repo/chief_complaint/chief_complaint_hpi_eye.go
package chief_complaint

import (
	"gorm.io/gorm"
	cc "sighthub-backend/internal/models/medical/vision_exam/chief_complaint"
)

type ChiefComplaintHPIEyeRepo struct{ DB *gorm.DB }

func NewChiefComplaintHPIEyeRepo(db *gorm.DB) *ChiefComplaintHPIEyeRepo {
	return &ChiefComplaintHPIEyeRepo{DB: db}
}

func (r *ChiefComplaintHPIEyeRepo) GetByID(id int64) (*cc.ChiefComplaintHPIEye, error) {
	var v cc.ChiefComplaintHPIEye
	if err := r.DB.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ChiefComplaintHPIEyeRepo) Create() (*cc.ChiefComplaintHPIEye, error) {
	v := cc.ChiefComplaintHPIEye{}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ChiefComplaintHPIEyeRepo) Save(v *cc.ChiefComplaintHPIEye) error {
	return r.DB.Save(v).Error
}

func (r *ChiefComplaintHPIEyeRepo) Delete(id int64) error {
	return r.DB.Delete(&cc.ChiefComplaintHPIEye{}, id).Error
}
