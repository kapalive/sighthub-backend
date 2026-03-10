package super_bill

import (
	"errors"

	"gorm.io/gorm"
	sb "sighthub-backend/internal/models/medical/vision_exam/super_bill"
)

type DiseaseSuperBillRepo struct{ DB *gorm.DB }

func NewDiseaseSuperBillRepo(db *gorm.DB) *DiseaseSuperBillRepo {
	return &DiseaseSuperBillRepo{DB: db}
}

func (r *DiseaseSuperBillRepo) GetByID(id int64) (*sb.DiseaseSuperBill, error) {
	var v sb.DiseaseSuperBill
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *DiseaseSuperBillRepo) Create(v *sb.DiseaseSuperBill) error {
	return r.DB.Create(v).Error
}

func (r *DiseaseSuperBillRepo) Save(v *sb.DiseaseSuperBill) error {
	return r.DB.Save(v).Error
}

func (r *DiseaseSuperBillRepo) Delete(id int64) error {
	return r.DB.Delete(&sb.DiseaseSuperBill{}, id).Error
}
