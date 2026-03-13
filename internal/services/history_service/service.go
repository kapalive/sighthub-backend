package history_service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	authModel   "sighthub-backend/internal/models/auth"
	empModel    "sighthub-backend/internal/models/employees"
	genModel    "sighthub-backend/internal/models/general"
	locModel    "sighthub-backend/internal/models/location"
	medModel    "sighthub-backend/internal/models/medical"
	patModel    "sighthub-backend/internal/models/patients"
	visionModel "sighthub-backend/internal/models/vision_exam"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── helpers ──────────────────────────────────────────────────────────────────

func (s *Service) getEmployee(username string) (*empModel.Employee, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("username = ?", username).First(&login).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	return &emp, nil
}

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, nil, err
	}
	if emp.LocationID == nil {
		return nil, nil, errors.New("location not found")
	}
	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, nil, errors.New("location not found")
	}
	return emp, &loc, nil
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ─── input types ──────────────────────────────────────────────────────────────

type OcularHistoryInput struct {
	PrevEyeHistoryNone *bool   `json:"prev_eye_history_none"`
	PrevEyeHistory     *string `json:"prev_eye_history"`
	LastExamNone       *bool   `json:"last_exam_none"`
	LastExam           *string `json:"last_exam"`
	ClCurrentWearNone  *bool   `json:"cl_current_wear_none"`
	ClCurrentWearScl   *bool   `json:"cl_current_wear_scl"`
	ClCurrentWearRgp   *bool   `json:"cl_current_wear_rgp"`
	ClCurrentWearOther *string `json:"cl_current_wear_other"`
	ModalityDaily      *bool   `json:"modality_daily"`
	ModalityBiweekly   *bool   `json:"modality_biweekly"`
	ModalityMonthly    *bool   `json:"modality_monthly"`
	ModalityAnnually   *bool   `json:"modality_annually"`
	ModalityOther      *string `json:"modality_other"`
	ModalitySolutions  *string `json:"modality_solutions"`
}

type SocialHistoryInput struct {
	AlertToTime  *bool   `json:"alert_to_time"`
	AlertToPlace *bool   `json:"alert_to_place"`
	AwareOfSelf  *bool   `json:"aware_of_self"`
	SitsUpright  *bool   `json:"sits_upright"`
	AlcoholUse   *string `json:"alcohol_use"`
	TobaccoUse   *string `json:"tobacco_use"`
}

type ROSMedicalHistoryInput struct {
	Eyes                bool    `json:"eyes"`
	EyesText            *string `json:"eyes_text"`
	General             bool    `json:"general"`
	GeneralText         *string `json:"general_text"`
	Genitourinary       bool    `json:"genitourinary"`
	GenitourinaryText   *string `json:"genitourinary_text"`
	Gastrointestinal    bool    `json:"gastrointestinal"`
	GastrointestinalText *string `json:"gastrointestinal_text"`
	Psychiatric         bool    `json:"psychiatric"`
	PsychiatricText     *string `json:"psychiatric_text"`
	Endocrine           bool    `json:"endocrine"`
	EndocrineText       *string `json:"endocrine_text"`
	EarNoseThroat       bool    `json:"ear_nose_throat"`
	EarNoseThroatText   *string `json:"ear_nose_throat_text"`
	AllergyImmun        bool    `json:"allergy_immun"`
	AllergyImmunText    *string `json:"allergy_immun_text"`
	Integumentary       bool    `json:"integumentary"`
	IntegumentaryText   *string `json:"integumentary_text"`
	Cardiovascular      bool    `json:"cardiovascular"`
	CardiovascularText  *string `json:"cardiovascular_text"`
	Musculoskeletal     bool    `json:"musculoskeletal"`
	MusculoskeletalText *string `json:"musculoskeletal_text"`
	Respiratory         bool    `json:"respiratory"`
	RespiratoryText     *string `json:"respiratory_text"`
	HematologicalLymp   bool    `json:"hematological_lymp"`
	HematologicalLympText *string `json:"hematological_lymp_text"`
	Neurological        bool    `json:"neurological"`
	NeurologicalText    *string `json:"neurological_text"`
}

type FamilyHistoryInput struct {
	Cataract            bool    `json:"cataract"`
	Glaucoma            bool    `json:"glaucoma"`
	MacularDegeneration bool    `json:"macular_degeneration"`
	Diabetes            bool    `json:"diabetes"`
	Hypertension        bool    `json:"hypertension"`
	Cancer              bool    `json:"cancer"`
	HeartDisease        bool    `json:"heart_disease"`
	Note                *string `json:"note"`
}

type MedicalRecordInput struct {
	Occupation            *string `json:"occupation"`
	PersistingPatientNote *string `json:"persisting_patient_note"`
	RaceID                *int64  `json:"race_id"`
	EthnicityID           *int64  `json:"ethnicity_id"`
	Gender                *string `json:"gender"`
}

type SaveHistoryInput struct {
	MedicalRecord             MedicalRecordInput     `json:"medical_record"`
	OcularHistory             OcularHistoryInput     `json:"ocular_history"`
	SocialHistory             SocialHistoryInput     `json:"social_history"`
	ROSMedicalHistory         ROSMedicalHistoryInput `json:"ros_medical_history"`
	FamilyHistory             FamilyHistoryInput     `json:"family_history"`
	PrimaryCarePhysician      string                 `json:"primary_care_physician"`
	Other1PrimaryCarePhysician *string               `json:"other_1_primary_care_physician"`
	Other2PrimaryCarePhysician *string               `json:"other_2_primary_care_physician"`
	Medication                *string                `json:"medication"`
	Allergy                   *string                `json:"allergy"`
	NoMedications             bool                   `json:"no_medications"`
	NoKnownAllergies          bool                   `json:"no_known_allergies"`
	LeadingWildcard           bool                   `json:"leading_wildcard"`
	SeeScannedDocumentsFolder bool                   `json:"see_scanned_documents_folder"`
	HistoryNote               *string                `json:"history_note"`
}

// ─── SaveHistory (POST) ───────────────────────────────────────────────────────

func (s *Service) SaveHistory(username string, examID int64, input SaveHistoryInput) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("you are not authorized to update this exam")
	}
	if exam.Passed {
		return nil, errors.New("cannot create history for a completed exam")
	}

	var patient patModel.Patient
	if err := s.db.First(&patient, exam.PatientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	// Create sub-records
	mr := medModel.MedicalRecord{
		Occupation:            input.MedicalRecord.Occupation,
		PersistingPatientNote: input.MedicalRecord.PersistingPatientNote,
	}
	if err := s.db.Create(&mr).Error; err != nil {
		return nil, errors.New("failed to create medical record")
	}

	oh := visionModel.OcularHistory{
		PrevEyeHistoryNone: input.OcularHistory.PrevEyeHistoryNone,
		PrevEyeHistory:     input.OcularHistory.PrevEyeHistory,
		LastExamNone:       input.OcularHistory.LastExamNone,
		LastExam:           input.OcularHistory.LastExam,
		ClCurrentWearNone:  input.OcularHistory.ClCurrentWearNone,
		ClCurrentWearScl:   input.OcularHistory.ClCurrentWearScl,
		ClCurrentWearRgp:   input.OcularHistory.ClCurrentWearRgp,
		ClCurrentWearOther: input.OcularHistory.ClCurrentWearOther,
		ModalityDaily:      input.OcularHistory.ModalityDaily,
		ModalityBiweekly:   input.OcularHistory.ModalityBiweekly,
		ModalityMonthly:    input.OcularHistory.ModalityMonthly,
		ModalityAnnually:   input.OcularHistory.ModalityAnnually,
		ModalityOther:      input.OcularHistory.ModalityOther,
		ModalitySolutions:  input.OcularHistory.ModalitySolutions,
	}
	if err := s.db.Create(&oh).Error; err != nil {
		return nil, errors.New("failed to create ocular history")
	}

	sh := visionModel.SocialHistory{
		AlertToTime:  input.SocialHistory.AlertToTime,
		AlertToPlace: input.SocialHistory.AlertToPlace,
		AwareOfSelf:  input.SocialHistory.AwareOfSelf,
		SitsUpright:  input.SocialHistory.SitsUpright,
		AlcoholUse:   input.SocialHistory.AlcoholUse,
		TobaccoUse:   input.SocialHistory.TobaccoUse,
	}
	if err := s.db.Create(&sh).Error; err != nil {
		return nil, errors.New("failed to create social history")
	}

	ros := visionModel.ROSMedicalHistory{
		Eyes:                  input.ROSMedicalHistory.Eyes,
		EyesText:              input.ROSMedicalHistory.EyesText,
		General:               input.ROSMedicalHistory.General,
		GeneralText:           input.ROSMedicalHistory.GeneralText,
		Genitourinary:         input.ROSMedicalHistory.Genitourinary,
		GenitourinaryText:     input.ROSMedicalHistory.GenitourinaryText,
		Gastrointestinal:      input.ROSMedicalHistory.Gastrointestinal,
		GastrointestinalText:  input.ROSMedicalHistory.GastrointestinalText,
		Psychiatric:           input.ROSMedicalHistory.Psychiatric,
		PsychiatricText:       input.ROSMedicalHistory.PsychiatricText,
		Endocrine:             input.ROSMedicalHistory.Endocrine,
		EndocrineText:         input.ROSMedicalHistory.EndocrineText,
		EarNoseThroat:         input.ROSMedicalHistory.EarNoseThroat,
		EarNoseThroatText:     input.ROSMedicalHistory.EarNoseThroatText,
		AllergyImmun:          input.ROSMedicalHistory.AllergyImmun,
		AllergyImmunText:      input.ROSMedicalHistory.AllergyImmunText,
		Integumentary:         input.ROSMedicalHistory.Integumentary,
		IntegumentaryText:     input.ROSMedicalHistory.IntegumentaryText,
		Cardiovascular:        input.ROSMedicalHistory.Cardiovascular,
		CardiovascularText:    input.ROSMedicalHistory.CardiovascularText,
		Musculoskeletal:       input.ROSMedicalHistory.Musculoskeletal,
		MusculoskeletalText:   input.ROSMedicalHistory.MusculoskeletalText,
		Respiratory:           input.ROSMedicalHistory.Respiratory,
		RespiratoryText:       input.ROSMedicalHistory.RespiratoryText,
		HematologicalLymp:     input.ROSMedicalHistory.HematologicalLymp,
		HematologicalLympText: input.ROSMedicalHistory.HematologicalLympText,
		Neurological:          input.ROSMedicalHistory.Neurological,
		NeurologicalText:      input.ROSMedicalHistory.NeurologicalText,
	}
	if err := s.db.Create(&ros).Error; err != nil {
		return nil, errors.New("failed to create ROS medical history")
	}

	fh := visionModel.FamilyHistory{
		Cataract:            input.FamilyHistory.Cataract,
		Glaucoma:            input.FamilyHistory.Glaucoma,
		MacularDegeneration: input.FamilyHistory.MacularDegeneration,
		Diabetes:            input.FamilyHistory.Diabetes,
		Hypertension:        input.FamilyHistory.Hypertension,
		Cancer:              input.FamilyHistory.Cancer,
		HeartDisease:        input.FamilyHistory.HeartDisease,
		Note:                input.FamilyHistory.Note,
	}
	if err := s.db.Create(&fh).Error; err != nil {
		return nil, errors.New("failed to create family history")
	}

	pcpStr := strPtrOrNil(input.PrimaryCarePhysician)
	var pcpVal string
	if pcpStr != nil {
		pcpVal = *pcpStr
	}
	he := visionModel.HistoryEye{
		EyeExamID:                 examID,
		MedicalRecordID:           mr.IDMedicalRecord,
		OcularHistoryID:           oh.IDOcularHistory,
		ROSMedicalHistoryID:       ros.IDROSMedicalHistory,
		FamilyHistoryID:           fh.IDFamilyHistory,
		SocialHistoryID:           sh.IDSocialHistory,
		PrimaryCarePhysician:      &pcpVal,
		Other1PrimaryCarePhysician: input.Other1PrimaryCarePhysician,
		Other2PrimaryCarePhysician: input.Other2PrimaryCarePhysician,
		Medication:                input.Medication,
		Allergy:                   input.Allergy,
		NoMedications:             input.NoMedications,
		NoKnownAllergies:          input.NoKnownAllergies,
		LeadingWildcard:           input.LeadingWildcard,
		SeeScannedDocumentsFolder: input.SeeScannedDocumentsFolder,
		HistoryNote:               input.HistoryNote,
	}
	if err := s.db.Create(&he).Error; err != nil {
		return nil, errors.New("failed to create history eye")
	}

	// Update patient demographics
	updates := map[string]interface{}{}
	if input.MedicalRecord.RaceID != nil && (patient.RaceID == nil || *input.MedicalRecord.RaceID != *patient.RaceID) {
		updates["race_id"] = *input.MedicalRecord.RaceID
	}
	if input.MedicalRecord.EthnicityID != nil && (patient.EthnicityID == nil || *input.MedicalRecord.EthnicityID != *patient.EthnicityID) {
		updates["ethnicity_id"] = *input.MedicalRecord.EthnicityID
	}
	if input.MedicalRecord.Gender != nil && string(patient.Gender) != *input.MedicalRecord.Gender {
		updates["gender"] = *input.MedicalRecord.Gender
	}
	if len(updates) > 0 {
		s.db.Model(&patient).Updates(updates)
	}

	return he.ToMap(), nil
}

// ─── GetHistory (GET) ─────────────────────────────────────────────────────────

func (s *Service) GetHistory(examID int64) (map[string]interface{}, error) {
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	var patient patModel.Patient
	if err := s.db.First(&patient, exam.PatientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	var he visionModel.HistoryEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&he).Error; err != nil {
		// Return empty template
		return map[string]interface{}{
			"exam_id": exam.IDEyeExam,
			"exists":  false,
			"passed":  exam.Passed,
			"medical_record": map[string]interface{}{
				"occupation": nil, "persisting_patient_note": nil,
				"gender": nil, "race": nil, "ethnicity": nil,
			},
			"ocular_history": map[string]interface{}{
				"prev_eye_history_none": nil, "prev_eye_history": nil,
				"last_exam_none": nil, "last_exam": nil,
				"cl_current_wear_none": nil, "cl_current_wear_scl": nil,
				"cl_current_wear_rgp": nil, "cl_current_wear_other": nil,
				"modality_daily": nil, "modality_biweekly": nil,
				"modality_monthly": nil, "modality_annually": nil,
				"modality_other": nil, "modality_solutions": nil,
			},
			"ros_medical_history": map[string]interface{}{
				"eyes": nil, "eyes_text": nil, "general": nil, "general_text": nil,
				"genitourinary": nil, "genitourinary_text": nil,
				"gastrointestinal": nil, "gastrointestinal_text": nil,
				"psychiatric": nil, "psychiatric_text": nil,
				"endocrine": nil, "endocrine_text": nil,
				"ear_nose_throat": nil, "ear_nose_throat_text": nil,
				"allergy_immun": nil, "allergy_immun_text": nil,
				"integumentary": nil, "integumentary_text": nil,
				"cardiovascular": nil, "cardiovascular_text": nil,
				"musculoskeletal": nil, "musculoskeletal_text": nil,
				"respiratory": nil, "respiratory_text": nil,
				"hematological_lymp": nil, "hematological_lymp_text": nil,
				"neurological": nil, "neurological_text": nil,
			},
			"family_history": map[string]interface{}{
				"cataract": nil, "glaucoma": nil, "macular_degeneration": nil,
				"diabetes": nil, "hypertension": nil, "cancer": nil,
				"heart_disease": nil, "note": nil,
			},
			"social_history": map[string]interface{}{
				"alert_to_time": nil, "alert_to_place": nil,
				"aware_of_self": nil, "sits_upright": nil,
				"alcohol_use": nil, "tobacco_use": nil,
			},
			"primary_care_physician": nil, "other_1_primary_care_physician": nil,
			"other_2_primary_care_physician": nil, "medication": nil, "allergy": nil,
			"no_medications": nil, "no_known_allergies": nil,
			"leading_wildcard": nil, "see_scanned_documents_folder": nil,
			"history_note": nil,
		}, nil
	}

	// Load sub-records
	var mr medModel.MedicalRecord
	s.db.First(&mr, he.MedicalRecordID)
	var oh visionModel.OcularHistory
	s.db.First(&oh, he.OcularHistoryID)
	var ros visionModel.ROSMedicalHistory
	s.db.First(&ros, he.ROSMedicalHistoryID)
	var fh visionModel.FamilyHistory
	s.db.First(&fh, he.FamilyHistoryID)
	var sh visionModel.SocialHistory
	s.db.First(&sh, he.SocialHistoryID)

	medRecordMap := map[string]interface{}{
		"occupation":              mr.Occupation,
		"persisting_patient_note": mr.PersistingPatientNote,
		"gender":                  patient.Gender,
		"race_id":                 patient.RaceID,
		"ethnicity_id":            patient.EthnicityID,
	}

	return map[string]interface{}{
		"exam_id":                        exam.IDEyeExam,
		"exists":                         true,
		"passed":                         exam.Passed,
		"medical_record":                 medRecordMap,
		"ocular_history":                 oh.ToMap(),
		"ros_medical_history":            ros.ToMap(),
		"family_history":                 fh.ToMap(),
		"social_history":                 sh.ToMap(),
		"primary_care_physician":         he.PrimaryCarePhysician,
		"other_1_primary_care_physician": he.Other1PrimaryCarePhysician,
		"other_2_primary_care_physician": he.Other2PrimaryCarePhysician,
		"medication":                     he.Medication,
		"allergy":                        he.Allergy,
		"no_medications":                 he.NoMedications,
		"no_known_allergies":             he.NoKnownAllergies,
		"leading_wildcard":               he.LeadingWildcard,
		"see_scanned_documents_folder":   he.SeeScannedDocumentsFolder,
		"history_note":                   he.HistoryNote,
	}, nil
}

// ─── UpdateHistory (PUT) ──────────────────────────────────────────────────────

type UpdateHistoryInput struct {
	MedicalRecord             *MedicalRecordInput     `json:"medical_record"`
	OcularHistory             *OcularHistoryInput     `json:"ocular_history"`
	SocialHistory             *SocialHistoryInput     `json:"social_history"`
	ROSMedicalHistory         *ROSMedicalHistoryInput `json:"ros_medical_history"`
	FamilyHistory             *FamilyHistoryInput     `json:"family_history"`
	PrimaryCarePhysician      *string                 `json:"primary_care_physician"`
	Other1PrimaryCarePhysician *string                `json:"other_1_primary_care_physician"`
	Other2PrimaryCarePhysician *string                `json:"other_2_primary_care_physician"`
	NoMedications             *bool                   `json:"no_medications"`
	NoKnownAllergies          *bool                   `json:"no_known_allergies"`
	LeadingWildcard           *bool                   `json:"leading_wildcard"`
	SeeScannedDocumentsFolder *bool                   `json:"see_scanned_documents_folder"`
	HistoryNote               *string                 `json:"history_note"`
}

func (s *Service) UpdateHistory(username string, examID int64, input UpdateHistoryInput) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.Passed {
		return nil, errors.New("cannot update history for a completed exam")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("you are not authorized to update this exam")
	}

	var he visionModel.HistoryEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&he).Error; err != nil {
		return nil, errors.New("history record not found for this exam")
	}

	heUpdates := map[string]interface{}{}

	if input.PrimaryCarePhysician != nil {
		heUpdates["primary_care_physician"] = *input.PrimaryCarePhysician
	}
	if input.Other1PrimaryCarePhysician != nil {
		heUpdates["other_1_primary_care_physician"] = *input.Other1PrimaryCarePhysician
	}
	if input.Other2PrimaryCarePhysician != nil {
		heUpdates["other_2_primary_care_physician"] = *input.Other2PrimaryCarePhysician
	}
	if input.NoMedications != nil {
		heUpdates["no_medications"] = *input.NoMedications
	}
	if input.NoKnownAllergies != nil {
		heUpdates["no_known_allergies"] = *input.NoKnownAllergies
	}
	if input.LeadingWildcard != nil {
		heUpdates["leading_wildcard"] = *input.LeadingWildcard
	}
	if input.SeeScannedDocumentsFolder != nil {
		heUpdates["see_scanned_documents_folder"] = *input.SeeScannedDocumentsFolder
	}
	if input.HistoryNote != nil {
		heUpdates["history_note"] = *input.HistoryNote
	}

	if input.OcularHistory != nil {
		oh := visionModel.OcularHistory{
			PrevEyeHistoryNone: input.OcularHistory.PrevEyeHistoryNone,
			PrevEyeHistory:     input.OcularHistory.PrevEyeHistory,
			LastExamNone:       input.OcularHistory.LastExamNone,
			LastExam:           input.OcularHistory.LastExam,
			ClCurrentWearNone:  input.OcularHistory.ClCurrentWearNone,
			ClCurrentWearScl:   input.OcularHistory.ClCurrentWearScl,
			ClCurrentWearRgp:   input.OcularHistory.ClCurrentWearRgp,
			ClCurrentWearOther: input.OcularHistory.ClCurrentWearOther,
			ModalityDaily:      input.OcularHistory.ModalityDaily,
			ModalityBiweekly:   input.OcularHistory.ModalityBiweekly,
			ModalityMonthly:    input.OcularHistory.ModalityMonthly,
			ModalityAnnually:   input.OcularHistory.ModalityAnnually,
			ModalityOther:      input.OcularHistory.ModalityOther,
			ModalitySolutions:  input.OcularHistory.ModalitySolutions,
		}
		if err := s.db.Create(&oh).Error; err == nil {
			heUpdates["ocular_history_id"] = oh.IDOcularHistory
		}
	}

	if input.SocialHistory != nil {
		sh := visionModel.SocialHistory{
			AlertToTime:  input.SocialHistory.AlertToTime,
			AlertToPlace: input.SocialHistory.AlertToPlace,
			AwareOfSelf:  input.SocialHistory.AwareOfSelf,
			SitsUpright:  input.SocialHistory.SitsUpright,
			AlcoholUse:   input.SocialHistory.AlcoholUse,
			TobaccoUse:   input.SocialHistory.TobaccoUse,
		}
		if err := s.db.Create(&sh).Error; err == nil {
			heUpdates["social_history_id"] = sh.IDSocialHistory
		}
	}

	if input.ROSMedicalHistory != nil {
		ros := visionModel.ROSMedicalHistory{
			Eyes: input.ROSMedicalHistory.Eyes, EyesText: input.ROSMedicalHistory.EyesText,
			General: input.ROSMedicalHistory.General, GeneralText: input.ROSMedicalHistory.GeneralText,
			Genitourinary: input.ROSMedicalHistory.Genitourinary, GenitourinaryText: input.ROSMedicalHistory.GenitourinaryText,
			Gastrointestinal: input.ROSMedicalHistory.Gastrointestinal, GastrointestinalText: input.ROSMedicalHistory.GastrointestinalText,
			Psychiatric: input.ROSMedicalHistory.Psychiatric, PsychiatricText: input.ROSMedicalHistory.PsychiatricText,
			Endocrine: input.ROSMedicalHistory.Endocrine, EndocrineText: input.ROSMedicalHistory.EndocrineText,
			EarNoseThroat: input.ROSMedicalHistory.EarNoseThroat, EarNoseThroatText: input.ROSMedicalHistory.EarNoseThroatText,
			AllergyImmun: input.ROSMedicalHistory.AllergyImmun, AllergyImmunText: input.ROSMedicalHistory.AllergyImmunText,
			Integumentary: input.ROSMedicalHistory.Integumentary, IntegumentaryText: input.ROSMedicalHistory.IntegumentaryText,
			Cardiovascular: input.ROSMedicalHistory.Cardiovascular, CardiovascularText: input.ROSMedicalHistory.CardiovascularText,
			Musculoskeletal: input.ROSMedicalHistory.Musculoskeletal, MusculoskeletalText: input.ROSMedicalHistory.MusculoskeletalText,
			Respiratory: input.ROSMedicalHistory.Respiratory, RespiratoryText: input.ROSMedicalHistory.RespiratoryText,
			HematologicalLymp: input.ROSMedicalHistory.HematologicalLymp, HematologicalLympText: input.ROSMedicalHistory.HematologicalLympText,
			Neurological: input.ROSMedicalHistory.Neurological, NeurologicalText: input.ROSMedicalHistory.NeurologicalText,
		}
		if err := s.db.Create(&ros).Error; err == nil {
			heUpdates["ros_medical_history_id"] = ros.IDROSMedicalHistory
		}
	}

	if input.FamilyHistory != nil {
		fh := visionModel.FamilyHistory{
			Cataract: input.FamilyHistory.Cataract, Glaucoma: input.FamilyHistory.Glaucoma,
			MacularDegeneration: input.FamilyHistory.MacularDegeneration,
			Diabetes: input.FamilyHistory.Diabetes, Hypertension: input.FamilyHistory.Hypertension,
			Cancer: input.FamilyHistory.Cancer, HeartDisease: input.FamilyHistory.HeartDisease,
			Note: input.FamilyHistory.Note,
		}
		if err := s.db.Create(&fh).Error; err == nil {
			heUpdates["family_history_id"] = fh.IDFamilyHistory
		}
	}

	// medical_record is required for update (like Python)
	if input.MedicalRecord == nil {
		return nil, errors.New("medical_record is required")
	}
	mr := medModel.MedicalRecord{
		Occupation:            input.MedicalRecord.Occupation,
		PersistingPatientNote: input.MedicalRecord.PersistingPatientNote,
	}
	if err := s.db.Create(&mr).Error; err != nil {
		return nil, errors.New("failed to create medical record")
	}
	heUpdates["medical_record_id"] = mr.IDMedicalRecord

	// Update patient demographics
	var patient patModel.Patient
	if err := s.db.First(&patient, exam.PatientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}
	patUpdates := map[string]interface{}{}
	if input.MedicalRecord.Gender != nil {
		patUpdates["gender"] = *input.MedicalRecord.Gender
	}
	if input.MedicalRecord.RaceID != nil {
		patUpdates["race_id"] = *input.MedicalRecord.RaceID
	}
	if input.MedicalRecord.EthnicityID != nil {
		patUpdates["ethnicity_id"] = *input.MedicalRecord.EthnicityID
	}
	if len(patUpdates) > 0 {
		s.db.Model(&patient).Updates(patUpdates)
	}

	if len(heUpdates) > 0 {
		if err := s.db.Model(&he).Updates(heUpdates).Error; err != nil {
			return nil, errors.New("failed to save data")
		}
	}

	return he.ToMap(), nil
}

// ─── Races / Ethnicities ──────────────────────────────────────────────────────

func (s *Service) GetRaces() ([]map[string]interface{}, error) {
	var races []genModel.Race
	if err := s.db.Find(&races).Error; err != nil {
		return nil, errors.New("failed to fetch races")
	}
	out := make([]map[string]interface{}, len(races))
	for i, r := range races {
		out[i] = r.ToMap()
	}
	return out, nil
}

func (s *Service) GetEthnicities() ([]map[string]interface{}, error) {
	var ethnicities []genModel.Ethnicity
	if err := s.db.Find(&ethnicities).Error; err != nil {
		return nil, errors.New("failed to fetch ethnicities")
	}
	out := make([]map[string]interface{}, len(ethnicities))
	for i, e := range ethnicities {
		out[i] = e.ToMap()
	}
	return out, nil
}

// ─── PatientInfo ──────────────────────────────────────────────────────────────

func (s *Service) GetPatientInfo(examID int64) (map[string]interface{}, error) {
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	var patient patModel.Patient
	if err := s.db.First(&patient, exam.PatientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	result := map[string]interface{}{
		"gender":    patient.Gender,
		"race":      nil,
		"ethnicity": nil,
	}
	if patient.RaceID != nil {
		var race genModel.Race
		if err := s.db.First(&race, *patient.RaceID).Error; err == nil {
			result["race"] = race.ToMap()
		}
	}
	if patient.EthnicityID != nil {
		var ethnicity genModel.Ethnicity
		if err := s.db.First(&ethnicity, *patient.EthnicityID).Error; err == nil {
			result["ethnicity"] = ethnicity.ToMap()
		}
	}
	return result, nil
}

// ─── Medications ──────────────────────────────────────────────────────────────

func (s *Service) SaveMedication(username string, examID int64, title string) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}
	_ = emp

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return errors.New("exam not found")
	}
	if exam.Passed {
		return errors.New("cannot modify a completed exam")
	}

	parts := strings.Split(title, "|")
	name := strings.TrimSpace(parts[0])
	var formulationType, strength *string
	if len(parts) > 1 {
		v := strings.TrimSpace(parts[1])
		if v != "" {
			formulationType = &v
		}
	}
	if len(parts) > 2 {
		v := strings.TrimSpace(parts[2])
		if v != "" {
			strength = &v
		}
	}

	med := medModel.UseMedications{
		Title:           name,
		FormulationType: formulationType,
		Strength:        strength,
		EyeExamID:       examID,
	}
	return s.db.Create(&med).Error
}

func (s *Service) GetMedications(username string, examID int64) ([]map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	_ = emp

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}

	var meds []medModel.UseMedications
	if err := s.db.Where("eye_exam_id = ?", examID).Find(&meds).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(meds))
	for i, m := range meds {
		out[i] = map[string]interface{}{
			"id_use_medications": m.IDUseMedications,
			"title":              m.Title,
			"formulation_type":   m.FormulationType,
			"strength":           m.Strength,
		}
	}
	return out, nil
}

func (s *Service) DeleteMedication(username string, examID, medicationID int64) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}
	_ = emp

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return errors.New("exam not found")
	}
	if exam.Passed {
		return errors.New("cannot modify a completed exam")
	}

	var med medModel.UseMedications
	if err := s.db.Where("id_use_medications = ? AND eye_exam_id = ?", medicationID, examID).
		First(&med).Error; err != nil {
		return errors.New("medication not found")
	}
	return s.db.Delete(&med).Error
}

// ─── Allergies ────────────────────────────────────────────────────────────────

func (s *Service) SaveAllergy(username string, examID int64, title string) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}
	_ = emp

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return errors.New("exam not found")
	}
	if exam.Passed {
		return errors.New("cannot modify a completed exam")
	}

	allergy := medModel.KnownAllergies{
		Title:     title,
		EyeExamID: examID,
	}
	return s.db.Create(&allergy).Error
}

func (s *Service) GetAllergies(username string, examID int64) ([]map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	_ = emp

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}

	var allergies []medModel.KnownAllergies
	if err := s.db.Where("eye_exam_id = ?", examID).Find(&allergies).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(allergies))
	for i, a := range allergies {
		out[i] = map[string]interface{}{
			"id_known_allergies": a.IDKnownAllergies,
			"title":              a.Title,
			"eye_exam_id":        a.EyeExamID,
		}
	}
	return out, nil
}

func (s *Service) DeleteAllergy(username string, examID, allergyID int64) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}
	_ = emp

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return errors.New("exam not found")
	}
	if exam.Passed {
		return errors.New("cannot modify a completed exam")
	}

	var allergy medModel.KnownAllergies
	if err := s.db.Where("id_known_allergies = ? AND eye_exam_id = ?", allergyID, examID).
		First(&allergy).Error; err != nil {
		return errors.New("allergy not found")
	}
	return s.db.Delete(&allergy).Error
}
