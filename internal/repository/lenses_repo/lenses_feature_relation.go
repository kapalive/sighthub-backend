package lenses_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensesFeatureRelationRepo struct{ DB *gorm.DB }

func NewLensesFeatureRelationRepo(db *gorm.DB) *LensesFeatureRelationRepo {
	return &LensesFeatureRelationRepo{DB: db}
}

func (r *LensesFeatureRelationRepo) GetByLensID(lensID int) ([]lenses.LensesFeatureRelation, error) {
	var items []lenses.LensesFeatureRelation
	return items, r.DB.Preload("LensSpecialFeature").Where("lenses_id = ?", lensID).Find(&items).Error
}

func (r *LensesFeatureRelationRepo) Add(lensID, featureID int) error {
	v := lenses.LensesFeatureRelation{LensesID: lensID, LensSpecialFeaturesID: featureID}
	return r.DB.Create(&v).Error
}

func (r *LensesFeatureRelationRepo) Remove(lensID, featureID int) error {
	return r.DB.Where("lenses_id = ? AND lens_special_features_id = ?", lensID, featureID).
		Delete(&lenses.LensesFeatureRelation{}).Error
}

func (r *LensesFeatureRelationRepo) RemoveAllForLens(lensID int) error {
	return r.DB.Where("lenses_id = ?", lensID).Delete(&lenses.LensesFeatureRelation{}).Error
}
