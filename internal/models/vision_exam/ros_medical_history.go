// internal/models/vision_exam/ros_medical_history.go
package vision_exam

// ROSMedicalHistory ↔ table: ros_medical_history (Review of Systems)
type ROSMedicalHistory struct {
	IDROSMedicalHistory   int64   `gorm:"column:id_ros_medical_history;primaryKey;autoIncrement" json:"id_ros_medical_history"`
	Eyes                  bool    `gorm:"column:eyes;not null"                                   json:"eyes"`
	EyesText              *string `gorm:"column:eyes_text;type:text"                             json:"eyes_text,omitempty"`
	General               bool    `gorm:"column:general;not null"                                json:"general"`
	GeneralText           *string `gorm:"column:general_text;type:text"                          json:"general_text,omitempty"`
	Genitourinary         bool    `gorm:"column:genitourinary;not null"                          json:"genitourinary"`
	GenitourinaryText     *string `gorm:"column:genitourinary_text;type:text"                    json:"genitourinary_text,omitempty"`
	Gastrointestinal      bool    `gorm:"column:gastrointestinal;not null"                       json:"gastrointestinal"`
	GastrointestinalText  *string `gorm:"column:gastrointestinal_text;type:text"                 json:"gastrointestinal_text,omitempty"`
	Psychiatric           bool    `gorm:"column:psychiatric;not null"                            json:"psychiatric"`
	PsychiatricText       *string `gorm:"column:psychiatric_text;type:text"                      json:"psychiatric_text,omitempty"`
	Endocrine             bool    `gorm:"column:endocrine;not null"                              json:"endocrine"`
	EndocrineText         *string `gorm:"column:endocrine_text;type:text"                        json:"endocrine_text,omitempty"`
	EarNoseThroat         bool    `gorm:"column:ear_nose_throat;not null"                        json:"ear_nose_throat"`
	EarNoseThroatText     *string `gorm:"column:ear_nose_throat_text;type:text"                  json:"ear_nose_throat_text,omitempty"`
	AllergyImmun          bool    `gorm:"column:allergy_immun;not null"                          json:"allergy_immun"`
	AllergyImmunText      *string `gorm:"column:allergy_immun_text;type:text"                    json:"allergy_immun_text,omitempty"`
	Integumentary         bool    `gorm:"column:integumentary;not null"                          json:"integumentary"`
	IntegumentaryText     *string `gorm:"column:integumentary_text;type:text"                    json:"integumentary_text,omitempty"`
	Cardiovascular        bool    `gorm:"column:cardiovascular;not null"                         json:"cardiovascular"`
	CardiovascularText    *string `gorm:"column:cardiovascular_text;type:text"                   json:"cardiovascular_text,omitempty"`
	Musculoskeletal       bool    `gorm:"column:musculoskeletal;not null"                        json:"musculoskeletal"`
	MusculoskeletalText   *string `gorm:"column:musculoskeletal_text;type:text"                  json:"musculoskeletal_text,omitempty"`
	Respiratory           bool    `gorm:"column:respiratory;not null"                            json:"respiratory"`
	RespiratoryText       *string `gorm:"column:respiratory_text;type:text"                      json:"respiratory_text,omitempty"`
	HematologicalLymp     bool    `gorm:"column:hematological_lymp;not null"                     json:"hematological_lymp"`
	HematologicalLympText *string `gorm:"column:hematological_lymp_text;type:text"               json:"hematological_lymp_text,omitempty"`
	Neurological          bool    `gorm:"column:neurological;not null"                           json:"neurological"`
	NeurologicalText      *string `gorm:"column:neurological_text;type:text"                     json:"neurological_text,omitempty"`
}

func (ROSMedicalHistory) TableName() string { return "ros_medical_history" }

func (r *ROSMedicalHistory) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_ros_medical_history":  r.IDROSMedicalHistory,
		"eyes":                    r.Eyes,
		"eyes_text":               r.EyesText,
		"general":                 r.General,
		"general_text":            r.GeneralText,
		"genitourinary":           r.Genitourinary,
		"genitourinary_text":      r.GenitourinaryText,
		"gastrointestinal":        r.Gastrointestinal,
		"gastrointestinal_text":   r.GastrointestinalText,
		"psychiatric":             r.Psychiatric,
		"psychiatric_text":        r.PsychiatricText,
		"endocrine":               r.Endocrine,
		"endocrine_text":          r.EndocrineText,
		"ear_nose_throat":         r.EarNoseThroat,
		"ear_nose_throat_text":    r.EarNoseThroatText,
		"allergy_immun":           r.AllergyImmun,
		"allergy_immun_text":      r.AllergyImmunText,
		"integumentary":           r.Integumentary,
		"integumentary_text":      r.IntegumentaryText,
		"cardiovascular":          r.Cardiovascular,
		"cardiovascular_text":     r.CardiovascularText,
		"musculoskeletal":         r.Musculoskeletal,
		"musculoskeletal_text":    r.MusculoskeletalText,
		"respiratory":             r.Respiratory,
		"respiratory_text":        r.RespiratoryText,
		"hematological_lymp":      r.HematologicalLymp,
		"hematological_lymp_text": r.HematologicalLympText,
		"neurological":            r.Neurological,
		"neurological_text":       r.NeurologicalText,
	}
}
