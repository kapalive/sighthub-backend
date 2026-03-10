// internal/models/vision_exam/external_sle.go
package vision_exam

import "time"

// PachExternalSle ↔ table: pach_external_sle
type PachExternalSle struct {
	IDPachExternalSle int64   `gorm:"column:id_pach_external_sle;primaryKey;autoIncrement" json:"id_pach_external_sle"`
	Od                *string `gorm:"column:od;type:varchar(255)" json:"od,omitempty"`
	Os                *string `gorm:"column:os;type:varchar(255)" json:"os,omitempty"`
}
func (PachExternalSle) TableName() string { return "pach_external_sle" }
func (p *PachExternalSle) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_pach_external_sle": p.IDPachExternalSle, "od": p.Od, "os": p.Os}
}

// GonioscopyExternalSle ↔ table: gonioscopy_external_sle
type GonioscopyExternalSle struct {
	IDGonioscopyExternalSle int64   `gorm:"column:id_gonioscopy_external_sle;primaryKey;autoIncrement" json:"id_gonioscopy_external_sle"`
	OdSup                   *string `gorm:"column:od_sup;type:varchar(255)" json:"od_sup,omitempty"`
	OsSup                   *string `gorm:"column:os_sup;type:varchar(255)" json:"os_sup,omitempty"`
	OdInf                   *string `gorm:"column:od_inf;type:varchar(255)" json:"od_inf,omitempty"`
	OsInf                   *string `gorm:"column:os_inf;type:varchar(255)" json:"os_inf,omitempty"`
	OdNasal                 *string `gorm:"column:od_nasal;type:varchar(255)" json:"od_nasal,omitempty"`
	OsNasal                 *string `gorm:"column:os_nasal;type:varchar(255)" json:"os_nasal,omitempty"`
	OdTemp                  *string `gorm:"column:od_temp;type:varchar(255)" json:"od_temp,omitempty"`
	OsTemp                  *string `gorm:"column:os_temp;type:varchar(255)" json:"os_temp,omitempty"`
}
func (GonioscopyExternalSle) TableName() string { return "gonioscopy_external_sle" }
func (g *GonioscopyExternalSle) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_gonioscopy_external_sle": g.IDGonioscopyExternalSle,
		"od_sup": g.OdSup, "os_sup": g.OsSup,
		"od_inf": g.OdInf, "os_inf": g.OsInf,
		"od_nasal": g.OdNasal, "os_nasal": g.OsNasal,
		"od_temp": g.OdTemp, "os_temp": g.OsTemp,
	}
}

// FindingsExternalSle ↔ table: findings_external_sle
type FindingsExternalSle struct {
	IDFindingsExternalSle      int64   `gorm:"column:id_findings_external_sle;primaryKey;autoIncrement" json:"id_findings_external_sle"`
	Externals                  *string `gorm:"column:externals;type:varchar(255)"                       json:"externals,omitempty"`
	OdLidsLashes               *string `gorm:"column:od_lids_lashes;type:varchar(255)"                  json:"od_lids_lashes,omitempty"`
	OsLidsLashes               *string `gorm:"column:os_lids_lashes;type:varchar(255)"                  json:"os_lids_lashes,omitempty"`
	OdConjunctivaSclera        *string `gorm:"column:od_conjunctiva_sclera;type:varchar(255)"           json:"od_conjunctiva_sclera,omitempty"`
	OsConjunctivaSclera        *string `gorm:"column:os_conjunctiva_sclera;type:varchar(255)"           json:"os_conjunctiva_sclera,omitempty"`
	OdCornea                   *string `gorm:"column:od_cornea;type:varchar(255)"                       json:"od_cornea,omitempty"`
	OsCornea                   *string `gorm:"column:os_cornea;type:varchar(255)"                       json:"os_cornea,omitempty"`
	OdTearFilm                 *string `gorm:"column:od_tear_film;type:varchar(255)"                    json:"od_tear_film,omitempty"`
	OsTearFilm                 *string `gorm:"column:os_tear_film;type:varchar(255)"                    json:"os_tear_film,omitempty"`
	OdAnteriorChamber          *string `gorm:"column:od_anterior_chamber;type:varchar(255)"             json:"od_anterior_chamber,omitempty"`
	OsAnteriorChamber          *string `gorm:"column:os_anterior_chamber;type:varchar(255)"             json:"os_anterior_chamber,omitempty"`
	OdIris                     *string `gorm:"column:od_iris;type:varchar(255)"                         json:"od_iris,omitempty"`
	OsIris                     *string `gorm:"column:os_iris;type:varchar(255)"                         json:"os_iris,omitempty"`
	OdLens                     *string `gorm:"column:od_lens;type:varchar(255)"                         json:"od_lens,omitempty"`
	OsLens                     *string `gorm:"column:os_lens;type:varchar(255)"                         json:"os_lens,omitempty"`
}
func (FindingsExternalSle) TableName() string { return "findings_external_sle" }
func (f *FindingsExternalSle) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_findings_external_sle": f.IDFindingsExternalSle,
		"externals": f.Externals,
		"od_lids_lashes": f.OdLidsLashes, "os_lids_lashes": f.OsLidsLashes,
		"od_conjunctiva_sclera": f.OdConjunctivaSclera, "os_conjunctiva_sclera": f.OsConjunctivaSclera,
		"od_cornea": f.OdCornea, "os_cornea": f.OsCornea,
		"od_tear_film": f.OdTearFilm, "os_tear_film": f.OsTearFilm,
		"od_anterior_chamber": f.OdAnteriorChamber, "os_anterior_chamber": f.OsAnteriorChamber,
		"od_iris": f.OdIris, "os_iris": f.OsIris,
		"od_lens": f.OdLens, "os_lens": f.OsLens,
	}
}

// VisualFields ↔ table: visual_fields
type VisualFields struct {
	IDVisualFields       int64   `gorm:"column:id_visual_fields;primaryKey;autoIncrement" json:"id_visual_fields"`
	SuperoTemporoOd      *string `gorm:"column:supero_temporo_od;type:varchar(255)"        json:"supero_temporo_od,omitempty"`
	SuperoTemporoOs      *string `gorm:"column:supero_temporo_os;type:varchar(255)"        json:"supero_temporo_os,omitempty"`
	SuperoNasalOd        *string `gorm:"column:supero_nasal_od;type:varchar(255)"          json:"supero_nasal_od,omitempty"`
	SuperoNasalOs        *string `gorm:"column:supero_nasal_os;type:varchar(255)"          json:"supero_nasal_os,omitempty"`
	InferoTemporalOd     *string `gorm:"column:infero_temporal_od;type:varchar(255)"       json:"infero_temporal_od,omitempty"`
	InferoTemporalOs     *string `gorm:"column:infero_temporal_os;type:varchar(255)"       json:"infero_temporal_os,omitempty"`
	InferoNasalOd        *string `gorm:"column:infero_nasal_od;type:varchar(255)"          json:"infero_nasal_od,omitempty"`
	InferoNasalOs        *string `gorm:"column:infero_nasal_os;type:varchar(255)"          json:"infero_nasal_os,omitempty"`
	Instrument           *string `gorm:"column:instrument;type:varchar(255)"               json:"instrument,omitempty"`
	Test                 *string `gorm:"column:test;type:varchar(255)"                     json:"test,omitempty"`
	Reason               *string `gorm:"column:reason;type:varchar(255)"                   json:"reason,omitempty"`
	Result               *string `gorm:"column:result;type:varchar(255)"                   json:"result,omitempty"`
	Recommendations      *string `gorm:"column:recommendations;type:varchar(255)"          json:"recommendations,omitempty"`
	Comments             *string `gorm:"column:comments;type:text"                         json:"comments,omitempty"`
}
func (VisualFields) TableName() string { return "visual_fields" }
func (v *VisualFields) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_visual_fields": v.IDVisualFields,
		"supero_temporo_od": v.SuperoTemporoOd, "supero_temporo_os": v.SuperoTemporoOs,
		"instrument": v.Instrument, "test": v.Test, "result": v.Result,
	}
}

// TonometryEye ↔ table: tonometry_eye
type TonometryEye struct {
	IDTonometryEye      int64      `gorm:"column:id_tonometry_eye;primaryKey;autoIncrement" json:"id_tonometry_eye"`
	ExternalSleEyeID    int64      `gorm:"column:external_sle_eye_id;not null"              json:"external_sle_eye_id"`
	MethodTonometry     *string    `gorm:"column:method_tonometry;type:varchar(100)"        json:"method_tonometry,omitempty"`
	DateTonometryEye    *time.Time `gorm:"column:date_tonometry_eye;type:date"              json:"date_tonometry_eye,omitempty"`
	TimeTonometryEye    *string    `gorm:"column:time_tonometry_eye;type:time"              json:"time_tonometry_eye,omitempty"`
	OdTonometryEye      *string    `gorm:"column:od_tonometry_eye;type:varchar(6)"          json:"od_tonometry_eye,omitempty"`
	OsTonometryEye      *string    `gorm:"column:os_tonometry_eye;type:varchar(6)"          json:"os_tonometry_eye,omitempty"`
}
func (TonometryEye) TableName() string { return "tonometry_eye" }
func (t *TonometryEye) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_tonometry_eye": t.IDTonometryEye, "external_sle_eye_id": t.ExternalSleEyeID,
		"method_tonometry": t.MethodTonometry, "od_tonometry_eye": t.OdTonometryEye,
		"os_tonometry_eye": t.OsTonometryEye, "time_tonometry_eye": t.TimeTonometryEye,
	}
	if t.DateTonometryEye != nil { m["date_tonometry_eye"] = t.DateTonometryEye.Format("2006-01-02") } else { m["date_tonometry_eye"] = nil }
	return m
}

// ExternalSleEye ↔ table: external_sle_eye
type ExternalSleEye struct {
	IDExternalSleEye        int64   `gorm:"column:id_external_sle_eye;primaryKey;autoIncrement"            json:"id_external_sle_eye"`
	FindingsExternalSleID   int64   `gorm:"column:findings_external_sle_id;not null;uniqueIndex"           json:"findings_external_sle_id"`
	AddDrawing              *string `gorm:"column:add_drawing;type:text"                                   json:"add_drawing,omitempty"`
	GonioscopyExternalSleID int64   `gorm:"column:gonioscopy_external_sle_id;not null;uniqueIndex"         json:"gonioscopy_external_sle_id"`
	PachExternalSleID       int64   `gorm:"column:pach_external_sle_id;not null;uniqueIndex"               json:"pach_external_sle_id"`
	OdAngleEstimation       string  `gorm:"column:od_angle_estimation;not null;default:'n/a'"              json:"od_angle_estimation"` // 1|2|3|4|n/a
	OsAngleEstimation       string  `gorm:"column:os_angle_estimation;not null;default:'n/a'"              json:"os_angle_estimation"`
	IopDropsFluress         bool    `gorm:"column:iop_drops_fluress;not null;default:false"                json:"iop_drops_fluress"`
	IopDropsProparacaine    bool    `gorm:"column:iop_drops_proparacaine;not null;default:false"           json:"iop_drops_proparacaine"`
	IopDropsFluoroStrip     bool    `gorm:"column:iop_drops_fluoro_strip;not null;default:false"           json:"iop_drops_fluoro_strip"`
	Note                    *string `gorm:"column:note;type:text"                                          json:"note,omitempty"`
	EyeExamID               int64   `gorm:"column:eye_exam_id;not null"                                    json:"eye_exam_id"`
	VisualFieldsID          *int64  `gorm:"column:visual_fields_id"                                        json:"visual_fields_id,omitempty"`

	FindingsExternalSle   *FindingsExternalSle   `gorm:"foreignKey:FindingsExternalSleID;references:IDFindingsExternalSle"     json:"-"`
	GonioscopyExternalSle *GonioscopyExternalSle `gorm:"foreignKey:GonioscopyExternalSleID;references:IDGonioscopyExternalSle" json:"-"`
	PachExternalSle       *PachExternalSle       `gorm:"foreignKey:PachExternalSleID;references:IDPachExternalSle"             json:"-"`
	VisualFields          *VisualFields          `gorm:"foreignKey:VisualFieldsID;references:IDVisualFields"                   json:"-"`
	TonometryReadings     []TonometryEye         `gorm:"foreignKey:ExternalSleEyeID;references:IDExternalSleEye"               json:"-"`
}
func (ExternalSleEye) TableName() string { return "external_sle_eye" }
func (e *ExternalSleEye) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_external_sle_eye": e.IDExternalSleEye,
		"findings_external_sle_id": e.FindingsExternalSleID,
		"gonioscopy_external_sle_id": e.GonioscopyExternalSleID,
		"pach_external_sle_id": e.PachExternalSleID,
		"od_angle_estimation": e.OdAngleEstimation, "os_angle_estimation": e.OsAngleEstimation,
		"iop_drops_fluress": e.IopDropsFluress, "iop_drops_proparacaine": e.IopDropsProparacaine,
		"iop_drops_fluoro_strip": e.IopDropsFluoroStrip,
		"note": e.Note, "eye_exam_id": e.EyeExamID, "visual_fields_id": e.VisualFieldsID,
	}
	if e.FindingsExternalSle != nil { m["findings_external_sle"] = e.FindingsExternalSle.ToMap() }
	if e.GonioscopyExternalSle != nil { m["gonioscopy_external_sle"] = e.GonioscopyExternalSle.ToMap() }
	if e.PachExternalSle != nil { m["pach_external_sle"] = e.PachExternalSle.ToMap() }
	if e.VisualFields != nil { m["visual_fields"] = e.VisualFields.ToMap() }
	return m
}
