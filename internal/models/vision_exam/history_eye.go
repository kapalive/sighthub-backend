// internal/models/vision_exam/history_eye.go
package vision_exam

import "sighthub-backend/internal/models/medical"

// HistoryEye ↔ table: history_eye (main history container per eye exam)
type HistoryEye struct {
	IDHistoryEye                  int64   `gorm:"column:id_history_eye;primaryKey;autoIncrement"                           json:"id_history_eye"`
	MedicalRecordID               int64   `gorm:"column:medical_record_id;not null;uniqueIndex"                            json:"medical_record_id"`
	OcularHistoryID               int64   `gorm:"column:ocular_history_id;not null;uniqueIndex"                            json:"ocular_history_id"`
	PrimaryCarePhysician          *string `gorm:"column:primary_care_physician;type:varchar(100)"                          json:"primary_care_physician,omitempty"`
	Other1PrimaryCarePhysician    *string `gorm:"column:other_1_primary_care_physician;type:varchar(100)"                  json:"other_1_primary_care_physician,omitempty"`
	Other2PrimaryCarePhysician    *string `gorm:"column:other_2_primary_care_physician;type:varchar(100)"                  json:"other_2_primary_care_physician,omitempty"`
	Medication                    *string `gorm:"column:medication;type:text"                                              json:"medication,omitempty"`
	Allergy                       *string `gorm:"column:allergy;type:text"                                                 json:"allergy,omitempty"`
	NoMedications                 bool    `gorm:"column:no_medications;not null;default:false"                             json:"no_medications"`
	NoKnownAllergies              bool    `gorm:"column:no_known_allergies;not null;default:false"                         json:"no_known_allergies"`
	LeadingWildcard               bool    `gorm:"column:leading_wildcard;not null;default:false"                           json:"leading_wildcard"`
	SeeScannedDocumentsFolder     bool    `gorm:"column:see_scanned_documents_folder;not null;default:false"               json:"see_scanned_documents_folder"`
	HistoryNote                   *string `gorm:"column:history_note;type:text"                                            json:"history_note,omitempty"`
	ROSMedicalHistoryID           int64   `gorm:"column:ros_medical_history_id;not null;uniqueIndex"                       json:"ros_medical_history_id"`
	FamilyHistoryID               int64   `gorm:"column:family_history_id;not null;uniqueIndex"                            json:"family_history_id"`
	SocialHistoryID               int64   `gorm:"column:social_history_id;not null;uniqueIndex"                            json:"social_history_id"`
	EyeExamID                     int64   `gorm:"column:eye_exam_id;not null"                                              json:"eye_exam_id"`

	// preload relations
	MedicalRecord    *medical.MedicalRecord `gorm:"foreignKey:MedicalRecordID;references:IDMedicalRecord"        json:"-"`
	OcularHistory    *OcularHistory         `gorm:"foreignKey:OcularHistoryID;references:IDOcularHistory"        json:"-"`
	ROSMedicalHistory *ROSMedicalHistory    `gorm:"foreignKey:ROSMedicalHistoryID;references:IDROSMedicalHistory" json:"-"`
	FamilyHistory    *FamilyHistory         `gorm:"foreignKey:FamilyHistoryID;references:IDFamilyHistory"        json:"-"`
	SocialHistory    *SocialHistory         `gorm:"foreignKey:SocialHistoryID;references:IDSocialHistory"        json:"-"`
}

func (HistoryEye) TableName() string { return "history_eye" }

func (h *HistoryEye) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_history_eye":                   h.IDHistoryEye,
		"medical_record_id":                h.MedicalRecordID,
		"ocular_history_id":                h.OcularHistoryID,
		"primary_care_physician":           h.PrimaryCarePhysician,
		"other_1_primary_care_physician":   h.Other1PrimaryCarePhysician,
		"other_2_primary_care_physician":   h.Other2PrimaryCarePhysician,
		"medication":                       h.Medication,
		"allergy":                          h.Allergy,
		"no_medications":                   h.NoMedications,
		"no_known_allergies":               h.NoKnownAllergies,
		"leading_wildcard":                 h.LeadingWildcard,
		"see_scanned_documents_folder":     h.SeeScannedDocumentsFolder,
		"history_note":                     h.HistoryNote,
		"ros_medical_history_id":           h.ROSMedicalHistoryID,
		"family_history_id":                h.FamilyHistoryID,
		"social_history_id":                h.SocialHistoryID,
		"eye_exam_id":                      h.EyeExamID,
	}
	if h.MedicalRecord != nil {
		m["medical_record"] = h.MedicalRecord.ToMap()
	}
	if h.OcularHistory != nil {
		m["ocular_history"] = h.OcularHistory.ToMap()
	}
	if h.ROSMedicalHistory != nil {
		m["ros_medical_history"] = h.ROSMedicalHistory.ToMap()
	}
	if h.FamilyHistory != nil {
		m["family_history"] = h.FamilyHistory.ToMap()
	}
	if h.SocialHistory != nil {
		m["social_history"] = h.SocialHistory.ToMap()
	}
	return m
}
