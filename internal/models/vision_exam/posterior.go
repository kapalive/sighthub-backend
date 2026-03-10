// internal/models/vision_exam/posterior.go
package vision_exam

import "time"

// FindingsPosterior ↔ table: findings_posterior
type FindingsPosterior struct {
	IDFindingsPosterior    int64   `gorm:"column:id_findings_posterior;primaryKey;autoIncrement" json:"id_findings_posterior"`
	OdView                 *string `gorm:"column:od_view;type:varchar(255)"           json:"od_view,omitempty"`
	OsView                 *string `gorm:"column:os_view;type:varchar(255)"           json:"os_view,omitempty"`
	OdVitreous             *string `gorm:"column:od_vitreous;type:varchar(255)"       json:"od_vitreous,omitempty"`
	OsVitreous             *string `gorm:"column:os_vitreous;type:varchar(255)"       json:"os_vitreous,omitempty"`
	OdMacula               *string `gorm:"column:od_macula;type:varchar(255)"         json:"od_macula,omitempty"`
	OsMacula               *string `gorm:"column:os_macula;type:varchar(255)"         json:"os_macula,omitempty"`
	OdBackground           *string `gorm:"column:od_background;type:varchar(255)"     json:"od_background,omitempty"`
	OsBackground           *string `gorm:"column:os_background;type:varchar(255)"     json:"os_background,omitempty"`
	OdVessels              *string `gorm:"column:od_vessels;type:varchar(255)"        json:"od_vessels,omitempty"`
	OsVessels              *string `gorm:"column:os_vessels;type:varchar(255)"        json:"os_vessels,omitempty"`
	OdPeripheralFundus     *string `gorm:"column:od_peripheral_fundus;type:varchar(255)" json:"od_peripheral_fundus,omitempty"`
	OsPeripheralFundus     *string `gorm:"column:os_peripheral_fundus;type:varchar(255)" json:"os_peripheral_fundus,omitempty"`
	OdOpticsNerve          *string `gorm:"column:od_optics_nerve;type:varchar(255)"   json:"od_optics_nerve,omitempty"`
	OsOpticsNerve          *string `gorm:"column:os_optics_nerve;type:varchar(255)"   json:"os_optics_nerve,omitempty"`
}
func (FindingsPosterior) TableName() string { return "findings_posterior" }
func (f *FindingsPosterior) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_findings_posterior": f.IDFindingsPosterior,
		"od_view": f.OdView, "os_view": f.OsView,
		"od_vitreous": f.OdVitreous, "os_vitreous": f.OsVitreous,
		"od_macula": f.OdMacula, "os_macula": f.OsMacula,
		"od_background": f.OdBackground, "os_background": f.OsBackground,
		"od_vessels": f.OdVessels, "os_vessels": f.OsVessels,
		"od_peripheral_fundus": f.OdPeripheralFundus, "os_peripheral_fundus": f.OsPeripheralFundus,
		"od_optics_nerve": f.OdOpticsNerve, "os_optics_nerve": f.OsOpticsNerve,
	}
}

// CupDiscRatioPosterior ↔ table: cup_disc_ratio_posterior
type CupDiscRatioPosterior struct {
	IDCupDiscRatioPosterior int64   `gorm:"column:id_cup_disc_ratio_posterior;primaryKey;autoIncrement" json:"id_cup_disc_ratio_posterior"`
	OdV                     *string `gorm:"column:od_v;type:varchar(255)" json:"od_v,omitempty"`
	OsV                     *string `gorm:"column:os_v;type:varchar(255)" json:"os_v,omitempty"`
	OdH                     *string `gorm:"column:od_h;type:varchar(255)" json:"od_h,omitempty"`
	OsH                     *string `gorm:"column:os_h;type:varchar(255)" json:"os_h,omitempty"`
}
func (CupDiscRatioPosterior) TableName() string { return "cup_disc_ratio_posterior" }
func (c *CupDiscRatioPosterior) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_cup_disc_ratio_posterior": c.IDCupDiscRatioPosterior,
		"od_v": c.OdV, "os_v": c.OsV, "od_h": c.OdH, "os_h": c.OsH,
	}
}

// PosteriorEye ↔ table: posterior_eye
type PosteriorEye struct {
	IDPosteriorEye              int64   `gorm:"column:id_posterior_eye;primaryKey;autoIncrement"        json:"id_posterior_eye"`
	InfoDirect                  bool    `gorm:"column:info_direct;not null;default:false"               json:"info_direct"`
	InfoBio                     bool    `gorm:"column:info_bio;not null;default:false"                  json:"info_bio"`
	Info90d                     bool    `gorm:"column:info_90d;not null;default:false"                  json:"info_90d"`
	InfoOptomap                 bool    `gorm:"column:info_optomap;not null;default:false"              json:"info_optomap"`
	InfoRha                     bool    `gorm:"column:info_rha;not null;default:false"                  json:"info_rha"`
	InfoOther                   *string `gorm:"column:info_other;type:text"                             json:"info_other,omitempty"`
	MedicationPatientEducated   bool    `gorm:"column:medication_patient_educated;not null;default:false" json:"medication_patient_educated"`
	MedicationDilationDeclined  bool    `gorm:"column:medication_ilation_declined;not null;default:false" json:"medication_ilation_declined"`
	MedicationParemyd           bool    `gorm:"column:medication_paremyd;not null;default:false"         json:"medication_paremyd"`
	MedicationAtropine          bool    `gorm:"column:medication_atropine;not null;default:false"        json:"medication_atropine"`
	MedicationTropicamide       bool    `gorm:"column:medication_tropicamide;not null;default:false"     json:"medication_tropicamide"`
	MedicationCyclopentolate    bool    `gorm:"column:medication_cyclopentolate;not null;default:false"  json:"medication_cyclopentolate"`
	MedicationHomatropine       bool    `gorm:"column:medication_homatropine;not null;default:false"     json:"medication_homatropine"`
	MedicationPhenylephrine     bool    `gorm:"column:medication_phenylephrine;not null;default:false"   json:"medication_phenylephrine"`
	MedicationRha               bool    `gorm:"column:medication_rha;not null;default:false"             json:"medication_rha"`
	TimeDilated                 *string `gorm:"column:time_dilated;type:time"                            json:"time_dilated,omitempty"`
	Other                       *string `gorm:"column:other;type:varchar(100)"                           json:"other,omitempty"`
	FindingsPosteriorID         int64   `gorm:"column:findings_posterior_id;not null;uniqueIndex"        json:"findings_posterior_id"`
	CupDiscRatioPosteriorID     int64   `gorm:"column:cup_disc_ratio_posterior_id;not null;uniqueIndex"  json:"cup_disc_ratio_posterior_id"`
	Note                        *string `gorm:"column:note;type:text"                                    json:"note,omitempty"`
	AddDrawing                  *string `gorm:"column:add_drawing;type:text"                             json:"add_drawing,omitempty"`
	EyeExamID                   int64   `gorm:"column:eye_exam_id;not null"                              json:"eye_exam_id"`

	FindingsPosterior     *FindingsPosterior     `gorm:"foreignKey:FindingsPosteriorID;references:IDFindingsPosterior"         json:"-"`
	CupDiscRatioPosterior *CupDiscRatioPosterior `gorm:"foreignKey:CupDiscRatioPosteriorID;references:IDCupDiscRatioPosterior" json:"-"`
}
func (PosteriorEye) TableName() string { return "posterior_eye" }
func (p *PosteriorEye) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_posterior_eye": p.IDPosteriorEye, "eye_exam_id": p.EyeExamID,
		"info_direct": p.InfoDirect, "info_bio": p.InfoBio, "info_90d": p.Info90d,
		"info_optomap": p.InfoOptomap, "info_rha": p.InfoRha, "info_other": p.InfoOther,
		"note": p.Note,
	}
	if p.FindingsPosterior != nil { m["findings_posterior"] = p.FindingsPosterior.ToMap() }
	if p.CupDiscRatioPosterior != nil { m["cup_disc_ratio_posterior"] = p.CupDiscRatioPosterior.ToMap() }
	return m
}

// Suppress unused import warning
var _ = time.Now
