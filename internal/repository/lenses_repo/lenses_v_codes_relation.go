package lenses_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensesVCodesRelationRepo struct{ DB *gorm.DB }

func NewLensesVCodesRelationRepo(db *gorm.DB) *LensesVCodesRelationRepo {
	return &LensesVCodesRelationRepo{DB: db}
}

func (r *LensesVCodesRelationRepo) GetByLensID(lensID int) ([]lenses.LensesVCodesRelation, error) {
	var items []lenses.LensesVCodesRelation
	return items, r.DB.Preload("VCode").Where("lenses_id = ?", lensID).Find(&items).Error
}

func (r *LensesVCodesRelationRepo) Add(lensID, vCodeID int) error {
	v := lenses.LensesVCodesRelation{LensesID: lensID, VCodesLensID: vCodeID}
	return r.DB.Create(&v).Error
}

func (r *LensesVCodesRelationRepo) Remove(lensID, vCodeID int) error {
	return r.DB.Where("lenses_id = ? AND v_codes_lens_id = ?", lensID, vCodeID).
		Delete(&lenses.LensesVCodesRelation{}).Error
}

func (r *LensesVCodesRelationRepo) RemoveAllForLens(lensID int) error {
	return r.DB.Where("lenses_id = ?", lensID).Delete(&lenses.LensesVCodesRelation{}).Error
}
