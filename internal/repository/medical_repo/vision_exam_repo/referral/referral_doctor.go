package referral

import (
	"errors"

	"gorm.io/gorm"
	ref "sighthub-backend/internal/models/medical/vision_exam/referral"
)

type ReferralDoctorRepo struct{ DB *gorm.DB }

func NewReferralDoctorRepo(db *gorm.DB) *ReferralDoctorRepo {
	return &ReferralDoctorRepo{DB: db}
}

func (r *ReferralDoctorRepo) GetByID(id int64) (*ref.ReferralDoctor, error) {
	var v ref.ReferralDoctor
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ReferralDoctorRepo) GetAll() ([]ref.ReferralDoctor, error) {
	var items []ref.ReferralDoctor
	if err := r.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ReferralDoctorRepo) Create(v *ref.ReferralDoctor) error {
	return r.DB.Create(v).Error
}

func (r *ReferralDoctorRepo) Save(v *ref.ReferralDoctor) error {
	return r.DB.Save(v).Error
}

func (r *ReferralDoctorRepo) Delete(id int64) error {
	return r.DB.Delete(&ref.ReferralDoctor{}, id).Error
}
