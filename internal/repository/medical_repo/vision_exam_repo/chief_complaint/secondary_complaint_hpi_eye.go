// internal/repository/medical_repo/vision_exam_repo/chief_complaint/secondary_complaint_hpi_eye.go
package chief_complaint

import (
	"gorm.io/gorm"
	cc "sighthub-backend/internal/models/medical/vision_exam/chief_complaint"
)

type SecondaryComplaintHPIEyeRepo struct{ DB *gorm.DB }

func NewSecondaryComplaintHPIEyeRepo(db *gorm.DB) *SecondaryComplaintHPIEyeRepo {
	return &SecondaryComplaintHPIEyeRepo{DB: db}
}

func (r *SecondaryComplaintHPIEyeRepo) GetByID(id int64) (*cc.SecondaryComplaintHPIEye, error) {
	var v cc.SecondaryComplaintHPIEye
	if err := r.DB.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SecondaryComplaintHPIEyeRepo) Create() (*cc.SecondaryComplaintHPIEye, error) {
	v := cc.SecondaryComplaintHPIEye{}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SecondaryComplaintHPIEyeRepo) Save(v *cc.SecondaryComplaintHPIEye) error {
	return r.DB.Save(v).Error
}

func (r *SecondaryComplaintHPIEyeRepo) Delete(id int64) error {
	return r.DB.Delete(&cc.SecondaryComplaintHPIEye{}, id).Error
}
