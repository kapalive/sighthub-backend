// internal/models/medical/vision_exam/chief_complaint/cc_hpi_eye.go
package chief_complaint

// CcHpiEye ↔ table: cc_hpi_eye
type CcHpiEye struct {
	IDCcHpiEye                 int64   `gorm:"column:id_cc_hpi_eye;primaryKey;autoIncrement"   json:"id_cc_hpi_eye"`
	ChiefComplaintHPIEyeID     *int64  `gorm:"column:chief_complaint_hpi_eye_id"               json:"chief_complaint_hpi_eye_id,omitempty"`
	ChiefComplaintNote         *string `gorm:"column:chief_complaint_note;type:text"           json:"chief_complaint_note,omitempty"`
	SecondaryComplaintHPIEyeID *int64  `gorm:"column:secondary_complaint_hpi_eye_id"           json:"secondary_complaint_hpi_eye_id,omitempty"`
	EyeExamID                  int64   `gorm:"column:eye_exam_id;not null"                     json:"eye_exam_id"`

	ChiefComplaint     *ChiefComplaintHPIEye     `gorm:"foreignKey:ChiefComplaintHPIEyeID;references:IDChiefComplaintHPIEye"         json:"-"`
	SecondaryComplaint *SecondaryComplaintHPIEye `gorm:"foreignKey:SecondaryComplaintHPIEyeID;references:IDSecondaryComplaintHPIEye" json:"-"`
}

func (CcHpiEye) TableName() string { return "cc_hpi_eye" }

func (c *CcHpiEye) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_cc_hpi_eye":                  c.IDCcHpiEye,
		"chief_complaint_hpi_eye_id":     c.ChiefComplaintHPIEyeID,
		"chief_complaint_note":           c.ChiefComplaintNote,
		"secondary_complaint_hpi_eye_id": c.SecondaryComplaintHPIEyeID,
		"eye_exam_id":                    c.EyeExamID,
	}
	if c.ChiefComplaint != nil {
		m["chief_complaint"] = c.ChiefComplaint.ToMap()
	}
	if c.SecondaryComplaint != nil {
		m["secondary_complaint"] = c.SecondaryComplaint.ToMap()
	}
	return m
}
