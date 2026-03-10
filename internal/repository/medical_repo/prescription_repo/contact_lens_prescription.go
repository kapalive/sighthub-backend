// internal/repository/medical_repo/prescription_repo/contact_lens_prescription.go
package prescription_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/prescriptions"
)

type ContactLensPrescriptionRepo struct{ DB *gorm.DB }

func NewContactLensPrescriptionRepo(db *gorm.DB) *ContactLensPrescriptionRepo {
	return &ContactLensPrescriptionRepo{DB: db}
}

func (r *ContactLensPrescriptionRepo) GetByPrescriptionID(prescriptionID int64) (*prescriptions.ContactLensPrescription, error) {
	var c prescriptions.ContactLensPrescription
	if err := r.DB.Where("prescription_id = ?", prescriptionID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

type CreateCLRxInput struct {
	PrescriptionID    int64
	OdContLens        *string
	OsContLens        *string
	OdBc              *string
	OsBc              *string
	OdDia             *float64
	OsDia             *float64
	OdPwr             *string
	OsPwr             *string
	OdCyl             *string
	OsCyl             *string
	OdAxis            *string
	OsAxis            *string
	OdAdd             *string
	OsAdd             *string
	OdColor           *string
	OsColor           *string
	OdType            *string
	OsType            *string
	ExpirationDate    *time.Time
	OdHPrismDirection *string
	OsHPrismDirection *string
	OdVPrismDirection *string
	OsVPrismDirection *string
}

func (r *ContactLensPrescriptionRepo) Create(inp CreateCLRxInput) (*prescriptions.ContactLensPrescription, error) {
	c := prescriptions.ContactLensPrescription{
		PrescriptionID:    inp.PrescriptionID,
		OdContLens:        inp.OdContLens,
		OsContLens:        inp.OsContLens,
		OdBc:              inp.OdBc,
		OsBc:              inp.OsBc,
		OdDia:             inp.OdDia,
		OsDia:             inp.OsDia,
		OdPwr:             inp.OdPwr,
		OsPwr:             inp.OsPwr,
		OdCyl:             inp.OdCyl,
		OsCyl:             inp.OsCyl,
		OdAxis:            inp.OdAxis,
		OsAxis:            inp.OsAxis,
		OdAdd:             inp.OdAdd,
		OsAdd:             inp.OsAdd,
		OdColor:           inp.OdColor,
		OsColor:           inp.OsColor,
		OdType:            inp.OdType,
		OsType:            inp.OsType,
		ExpirationDate:    inp.ExpirationDate,
		OdHPrismDirection: inp.OdHPrismDirection,
		OsHPrismDirection: inp.OsHPrismDirection,
		OdVPrismDirection: inp.OdVPrismDirection,
		OsVPrismDirection: inp.OsVPrismDirection,
	}
	if err := r.DB.Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ContactLensPrescriptionRepo) Save(c *prescriptions.ContactLensPrescription) error {
	return r.DB.Save(c).Error
}

func (r *ContactLensPrescriptionRepo) Delete(id int64) error {
	return r.DB.Delete(&prescriptions.ContactLensPrescription{}, id).Error
}

func (r *ContactLensPrescriptionRepo) DeleteByPrescriptionID(prescriptionID int64) error {
	return r.DB.Where("prescription_id = ?", prescriptionID).Delete(&prescriptions.ContactLensPrescription{}).Error
}
