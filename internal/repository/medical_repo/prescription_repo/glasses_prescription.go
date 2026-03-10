// internal/repository/medical_repo/prescription_repo/glasses_prescription.go
package prescription_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/prescriptions"
)

type GlassesPrescriptionRepo struct{ DB *gorm.DB }

func NewGlassesPrescriptionRepo(db *gorm.DB) *GlassesPrescriptionRepo {
	return &GlassesPrescriptionRepo{DB: db}
}

func (r *GlassesPrescriptionRepo) GetByPrescriptionID(prescriptionID int64) (*prescriptions.GlassesPrescription, error) {
	var g prescriptions.GlassesPrescription
	if err := r.DB.Where("prescription_id = ?", prescriptionID).First(&g).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

type CreateGlassesRxInput struct {
	PrescriptionID    int64
	OdSph             *string
	OsSph             *string
	OdCyl             *string
	OsCyl             *string
	OdAxis            *string
	OsAxis            *string
	OdAdd             *float64
	OsAdd             *float64
	OdHPrism          *float64
	OsHPrism          *float64
	OdHPrismDirection *string
	OsHPrismDirection *string
	OdVPrism          *float64
	OsVPrism          *float64
	OdVPrismDirection *string
	OsVPrismDirection *string
	OdDpd             *float64
	OsDpd             *float64
	ExpirationDate    *time.Time
}

func (r *GlassesPrescriptionRepo) Create(inp CreateGlassesRxInput) (*prescriptions.GlassesPrescription, error) {
	g := prescriptions.GlassesPrescription{
		PrescriptionID:    inp.PrescriptionID,
		OdSph:             inp.OdSph,
		OsSph:             inp.OsSph,
		OdCyl:             inp.OdCyl,
		OsCyl:             inp.OsCyl,
		OdAxis:            inp.OdAxis,
		OsAxis:            inp.OsAxis,
		OdAdd:             inp.OdAdd,
		OsAdd:             inp.OsAdd,
		OdHPrism:          inp.OdHPrism,
		OsHPrism:          inp.OsHPrism,
		OdHPrismDirection: inp.OdHPrismDirection,
		OsHPrismDirection: inp.OsHPrismDirection,
		OdVPrism:          inp.OdVPrism,
		OsVPrism:          inp.OsVPrism,
		OdVPrismDirection: inp.OdVPrismDirection,
		OsVPrismDirection: inp.OsVPrismDirection,
		OdDpd:             inp.OdDpd,
		OsDpd:             inp.OsDpd,
		ExpirationDate:    inp.ExpirationDate,
	}
	if err := r.DB.Create(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GlassesPrescriptionRepo) Save(g *prescriptions.GlassesPrescription) error {
	return r.DB.Save(g).Error
}

func (r *GlassesPrescriptionRepo) Delete(id int64) error {
	return r.DB.Delete(&prescriptions.GlassesPrescription{}, id).Error
}

func (r *GlassesPrescriptionRepo) DeleteByPrescriptionID(prescriptionID int64) error {
	return r.DB.Where("prescription_id = ?", prescriptionID).Delete(&prescriptions.GlassesPrescription{}).Error
}
