// internal/repository/medical_repo/prescription_repo/patient_prescription.go
package prescription_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/prescriptions"
)

type PatientPrescriptionRepo struct{ DB *gorm.DB }

func NewPatientPrescriptionRepo(db *gorm.DB) *PatientPrescriptionRepo {
	return &PatientPrescriptionRepo{DB: db}
}

func (r *PatientPrescriptionRepo) GetByPatientID(patientID int64) ([]prescriptions.PatientPrescription, error) {
	var list []prescriptions.PatientPrescription
	if err := r.DB.
		Preload("GlassesPrescription").
		Preload("ContactLensPrescription").
		Where("patient_id = ?", patientID).
		Order("id_patient_prescription DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PatientPrescriptionRepo) GetByID(id int64) (*prescriptions.PatientPrescription, error) {
	var p prescriptions.PatientPrescription
	if err := r.DB.
		Preload("GlassesPrescription").
		Preload("ContactLensPrescription").
		First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

type CreatePrescriptionInput struct {
	PatientID        int64
	PrescriptionDate *time.Time
	Note             *string
	Doctor           string
	NPI              *string
	License          *string
	LocationID       int
	Signature        *string
	Medication       *string
	Dosage           *string
	DocumentLink     *string
	GOrC             *string
	PhoneNumber      *string
}

func (r *PatientPrescriptionRepo) Create(inp CreatePrescriptionInput) (*prescriptions.PatientPrescription, error) {
	p := prescriptions.PatientPrescription{
		PatientID:        inp.PatientID,
		PrescriptionDate: inp.PrescriptionDate,
		Note:             inp.Note,
		Doctor:           inp.Doctor,
		NPI:              inp.NPI,
		License:          inp.License,
		LocationID:       inp.LocationID,
		Signature:        inp.Signature,
		Medication:       inp.Medication,
		Dosage:           inp.Dosage,
		DocumentLink:     inp.DocumentLink,
		GOrC:             inp.GOrC,
		PhoneNumber:      inp.PhoneNumber,
	}
	if err := r.DB.Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

type UpdatePrescriptionInput struct {
	PrescriptionDate *time.Time
	Note             *string
	Doctor           *string
	NPI              *string
	License          *string
	Signature        *string
	Medication       *string
	Dosage           *string
	DocumentLink     *string
	GOrC             *string
	PhoneNumber      *string
}

func (r *PatientPrescriptionRepo) Update(id int64, inp UpdatePrescriptionInput) error {
	updates := map[string]interface{}{}
	if inp.PrescriptionDate != nil {
		updates["prescription_date"] = inp.PrescriptionDate
	}
	if inp.Note != nil {
		updates["note"] = inp.Note
	}
	if inp.Doctor != nil {
		updates["doctor"] = *inp.Doctor
	}
	if inp.NPI != nil {
		updates["npi"] = inp.NPI
	}
	if inp.License != nil {
		updates["license"] = inp.License
	}
	if inp.Signature != nil {
		updates["signature"] = inp.Signature
	}
	if inp.Medication != nil {
		updates["medication"] = inp.Medication
	}
	if inp.Dosage != nil {
		updates["dosage"] = inp.Dosage
	}
	if inp.DocumentLink != nil {
		updates["document_link"] = inp.DocumentLink
	}
	if inp.GOrC != nil {
		updates["g_or_c"] = inp.GOrC
	}
	if inp.PhoneNumber != nil {
		updates["phone_number"] = inp.PhoneNumber
	}
	return r.DB.Model(&prescriptions.PatientPrescription{}).Where("id_patient_prescription = ?", id).Updates(updates).Error
}

func (r *PatientPrescriptionRepo) Delete(id int64) error {
	return r.DB.Delete(&prescriptions.PatientPrescription{}, id).Error
}
