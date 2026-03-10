package super_bill

import (
	"errors"

	"gorm.io/gorm"
	sb "sighthub-backend/internal/models/medical/vision_exam/super_bill"
)

type SuperBillDiagnosisRepo struct{ DB *gorm.DB }

func NewSuperBillDiagnosisRepo(db *gorm.DB) *SuperBillDiagnosisRepo {
	return &SuperBillDiagnosisRepo{DB: db}
}

func (r *SuperBillDiagnosisRepo) GetByID(id int64) (*sb.SuperBillDiagnosis, error) {
	var v sb.SuperBillDiagnosis
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *SuperBillDiagnosisRepo) Create(v *sb.SuperBillDiagnosis) error {
	return r.DB.Create(v).Error
}

func (r *SuperBillDiagnosisRepo) Save(v *sb.SuperBillDiagnosis) error {
	return r.DB.Save(v).Error
}

func (r *SuperBillDiagnosisRepo) Delete(id int64) error {
	return r.DB.Delete(&sb.SuperBillDiagnosis{}, id).Error
}
