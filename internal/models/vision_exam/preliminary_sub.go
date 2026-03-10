// internal/models/vision_exam/preliminary_sub.go
// Support tables for preliminary_eye_exam.
package vision_exam

import "time"

// EntranceGlasses ↔ table: entrance_glasses
type EntranceGlasses struct {
	IDEntranceGlasses int64      `gorm:"column:id_entrance_glasses;primaryKey;autoIncrement" json:"id_entrance_glasses"`
	Data              *time.Time `gorm:"column:data;type:date"                               json:"data,omitempty"`
	OdSph             *string    `gorm:"column:od_sph;type:varchar(10)"                      json:"od_sph,omitempty"`
	OsSph             *string    `gorm:"column:os_sph;type:varchar(10)"                      json:"os_sph,omitempty"`
	OdCyl             *string    `gorm:"column:od_cyl;type:varchar(10)"                      json:"od_cyl,omitempty"`
	OsCyl             *string    `gorm:"column:os_cyl;type:varchar(10)"                      json:"os_cyl,omitempty"`
	OdAxis            *string    `gorm:"column:od_axis;type:varchar(10)"                     json:"od_axis,omitempty"`
	OsAxis            *string    `gorm:"column:os_axis;type:varchar(10)"                     json:"os_axis,omitempty"`
	OdAdd             *string    `gorm:"column:od_add;type:varchar(10)"                      json:"od_add,omitempty"`
	OsAdd             *string    `gorm:"column:os_add;type:varchar(10)"                      json:"os_add,omitempty"`
	OdHPrism          *string    `gorm:"column:od_h_prism;type:varchar(10)"                  json:"od_h_prism,omitempty"`
	OsHPrism          *string    `gorm:"column:os_h_prism;type:varchar(10)"                  json:"os_h_prism,omitempty"`
	OdVPrism          *string    `gorm:"column:od_v_prism;type:varchar(10)"                  json:"od_v_prism,omitempty"`
	OsVPrism          *string    `gorm:"column:os_v_prism;type:varchar(10)"                  json:"os_v_prism,omitempty"`
	Note              *string    `gorm:"column:note;type:text"                               json:"note,omitempty"`
}

func (EntranceGlasses) TableName() string { return "entrance_glasses" }
func (e *EntranceGlasses) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_entrance_glasses": e.IDEntranceGlasses,
		"od_sph": e.OdSph, "os_sph": e.OsSph,
		"od_cyl": e.OdCyl, "os_cyl": e.OsCyl,
		"od_axis": e.OdAxis, "os_axis": e.OsAxis,
		"od_add": e.OdAdd, "os_add": e.OsAdd,
		"od_h_prism": e.OdHPrism, "os_h_prism": e.OsHPrism,
		"od_v_prism": e.OdVPrism, "os_v_prism": e.OsVPrism,
		"note": e.Note,
	}
	if e.Data != nil {
		m["data"] = e.Data.Format("2006-01-02")
	} else {
		m["data"] = nil
	}
	return m
}

// EntranceContLens ↔ table: entrance_cont_lens
type EntranceContLens struct {
	IDEntranceContLens int64      `gorm:"column:id_entrance_cont_lens;primaryKey;autoIncrement" json:"id_entrance_cont_lens"`
	Data               *time.Time `gorm:"column:data;type:date"                                 json:"data,omitempty"`
	OdBrand            *string    `gorm:"column:od_brand;type:varchar(100)"                     json:"od_brand,omitempty"`
	OsBrand            *string    `gorm:"column:os_brand;type:varchar(100)"                     json:"os_brand,omitempty"`
	OdBaseC            *string    `gorm:"column:od_base_c;type:varchar(50)"                     json:"od_base_c,omitempty"`
	OsBaseC            *string    `gorm:"column:os_base_c;type:varchar(50)"                     json:"os_base_c,omitempty"`
	OdDia              *string    `gorm:"column:od_dia;type:varchar(50)"                        json:"od_dia,omitempty"`
	OsDia              *string    `gorm:"column:os_dia;type:varchar(50)"                        json:"os_dia,omitempty"`
	OdPwr              *string    `gorm:"column:od_pwr;type:varchar(50)"                        json:"od_pwr,omitempty"`
	OsPwr              *string    `gorm:"column:os_pwr;type:varchar(50)"                        json:"os_pwr,omitempty"`
	OdCyl              *string    `gorm:"column:od_cyl;type:varchar(50)"                        json:"od_cyl,omitempty"`
	OsCyl              *string    `gorm:"column:os_cyl;type:varchar(50)"                        json:"os_cyl,omitempty"`
	OdAxis             *string    `gorm:"column:od_axis;type:varchar(50)"                       json:"od_axis,omitempty"`
	OsAxis             *string    `gorm:"column:os_axis;type:varchar(50)"                       json:"os_axis,omitempty"`
	OdAdd              *string    `gorm:"column:od_add;type:varchar(50)"                        json:"od_add,omitempty"`
	OsAdd              *string    `gorm:"column:os_add;type:varchar(50)"                        json:"os_add,omitempty"`
}

func (EntranceContLens) TableName() string { return "entrance_cont_lens" }
func (e *EntranceContLens) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_entrance_cont_lens": e.IDEntranceContLens,
		"od_brand": e.OdBrand, "os_brand": e.OsBrand,
		"od_base_c": e.OdBaseC, "os_base_c": e.OsBaseC,
		"od_dia": e.OdDia, "os_dia": e.OsDia,
		"od_pwr": e.OdPwr, "os_pwr": e.OsPwr,
		"od_cyl": e.OdCyl, "os_cyl": e.OsCyl,
		"od_axis": e.OdAxis, "os_axis": e.OsAxis,
		"od_add": e.OdAdd, "os_add": e.OsAdd,
	}
	if e.Data != nil {
		m["data"] = e.Data.Format("2006-01-02")
	} else {
		m["data"] = nil
	}
	return m
}

// UnaidedVADistance ↔ table: unaided_va_distance
type UnaidedVADistance struct {
	IDUnaidedVADistance int64   `gorm:"column:id_unaided_va_distance;primaryKey;autoIncrement" json:"id_unaided_va_distance"`
	Od20                *string `gorm:"column:od_20;type:varchar(255)"                         json:"od_20,omitempty"`
	Os20                *string `gorm:"column:os_20;type:varchar(255)"                         json:"os_20,omitempty"`
	Ou20                *string `gorm:"column:ou_20;type:varchar(255)"                         json:"ou_20,omitempty"`
}
func (UnaidedVADistance) TableName() string { return "unaided_va_distance" }
func (u *UnaidedVADistance) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_unaided_va_distance": u.IDUnaidedVADistance, "od_20": u.Od20, "os_20": u.Os20, "ou_20": u.Ou20}
}

// UnaidedPHDistance ↔ table: unaided_ph_distance
type UnaidedPHDistance struct {
	IDUnaidedPHDistance int64   `gorm:"column:id_unaided_ph_distance;primaryKey;autoIncrement" json:"id_unaided_ph_distance"`
	Od20                *string `gorm:"column:od_20;type:varchar(255)"                         json:"od_20,omitempty"`
	Os20                *string `gorm:"column:os_20;type:varchar(255)"                         json:"os_20,omitempty"`
}
func (UnaidedPHDistance) TableName() string { return "unaided_ph_distance" }
func (u *UnaidedPHDistance) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_unaided_ph_distance": u.IDUnaidedPHDistance, "od_20": u.Od20, "os_20": u.Os20}
}

// UnaidedVANear ↔ table: unaided_va_near
type UnaidedVANear struct {
	IDUnaidedVANear int64   `gorm:"column:id_unaided_va_near;primaryKey;autoIncrement" json:"id_unaided_va_near"`
	Od20            *string `gorm:"column:od_20;type:varchar(255)"                     json:"od_20,omitempty"`
	Os20            *string `gorm:"column:os_20;type:varchar(255)"                     json:"os_20,omitempty"`
	Ou20            *string `gorm:"column:ou_20;type:varchar(255)"                     json:"ou_20,omitempty"`
}
func (UnaidedVANear) TableName() string { return "unaided_va_near" }
func (u *UnaidedVANear) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_unaided_va_near": u.IDUnaidedVANear, "od_20": u.Od20, "os_20": u.Os20, "ou_20": u.Ou20}
}

// AidedVADistance ↔ table: aided_va_distance
type AidedVADistance struct {
	IDAidedVADistance int64   `gorm:"column:id_aided_va_distance;primaryKey;autoIncrement" json:"id_aided_va_distance"`
	Od20              *string `gorm:"column:od_20;type:varchar(255)"                       json:"od_20,omitempty"`
	Os20              *string `gorm:"column:os_20;type:varchar(255)"                       json:"os_20,omitempty"`
	Ou20              *string `gorm:"column:ou_20;type:varchar(255)"                       json:"ou_20,omitempty"`
}
func (AidedVADistance) TableName() string { return "aided_va_distance" }
func (a *AidedVADistance) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_aided_va_distance": a.IDAidedVADistance, "od_20": a.Od20, "os_20": a.Os20, "ou_20": a.Ou20}
}

// AidedPHDistance ↔ table: aided_ph_distance
type AidedPHDistance struct {
	IDAidedPHDistance int64   `gorm:"column:id_aided_ph_distance;primaryKey;autoIncrement" json:"id_aided_ph_distance"`
	Od20              *string `gorm:"column:od_20;type:varchar(255)"                       json:"od_20,omitempty"`
	Os20              *string `gorm:"column:os_20;type:varchar(255)"                       json:"os_20,omitempty"`
}
func (AidedPHDistance) TableName() string { return "aided_ph_distance" }
func (a *AidedPHDistance) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_aided_ph_distance": a.IDAidedPHDistance, "od_20": a.Od20, "os_20": a.Os20}
}

// AidedVANear ↔ table: aided_va_near
type AidedVANear struct {
	IDAidedVANear int64   `gorm:"column:id_aided_va_near;primaryKey;autoIncrement" json:"id_aided_va_near"`
	Od20          *string `gorm:"column:od_20;type:varchar(255)"                   json:"od_20,omitempty"`
	Os20          *string `gorm:"column:os_20;type:varchar(255)"                   json:"os_20,omitempty"`
	Ou20          *string `gorm:"column:ou_20;type:varchar(255)"                   json:"ou_20,omitempty"`
}
func (AidedVANear) TableName() string { return "aided_va_near" }
func (a *AidedVANear) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_aided_va_near": a.IDAidedVANear, "od_20": a.Od20, "os_20": a.Os20, "ou_20": a.Ou20}
}

// Confrontation ↔ table: confrontation
type Confrontation struct {
	IDConfrontation int64   `gorm:"column:id_confrontation;primaryKey;autoIncrement" json:"id_confrontation"`
	Od              *string `gorm:"column:od;type:varchar(255)"                      json:"od,omitempty"`
	Os              *string `gorm:"column:os;type:varchar(255)"                      json:"os,omitempty"`
}
func (Confrontation) TableName() string { return "confrontation" }
func (c *Confrontation) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_confrontation": c.IDConfrontation, "od": c.Od, "os": c.Os}
}

// Automated ↔ table: automated
type Automated struct {
	IDAutomated int64   `gorm:"column:id_automated;primaryKey;autoIncrement" json:"id_automated"`
	Od          *string `gorm:"column:od;type:varchar(255)"                  json:"od,omitempty"`
	Os          *string `gorm:"column:os;type:varchar(255)"                  json:"os,omitempty"`
}
func (Automated) TableName() string { return "automated" }
func (a *Automated) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_automated": a.IDAutomated, "od": a.Od, "os": a.Os}
}

// Motility ↔ table: motility
type Motility struct {
	IDMotility int64   `gorm:"column:id_motility;primaryKey;autoIncrement" json:"id_motility"`
	Od         *string `gorm:"column:od;type:varchar(255)"                 json:"od,omitempty"`
	Os         *string `gorm:"column:os;type:varchar(255)"                 json:"os,omitempty"`
}
func (Motility) TableName() string { return "motility" }
func (m *Motility) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_motility": m.IDMotility, "od": m.Od, "os": m.Os}
}

// Pupils ↔ table: pupils
type Pupils struct {
	IDPupils    int64   `gorm:"column:id_pupils;primaryKey;autoIncrement"  json:"id_pupils"`
	OdMmDim     *string `gorm:"column:od_mm_dim;type:varchar(255)"          json:"od_mm_dim,omitempty"`
	OdMmBright  *string `gorm:"column:od_mm_bright;type:varchar(255)"       json:"od_mm_bright,omitempty"`
	OsMmDim     *string `gorm:"column:os_mm_dim;type:varchar(255)"          json:"os_mm_dim,omitempty"`
	OsMmBright  *string `gorm:"column:os_mm_bright;type:varchar(255)"       json:"os_mm_bright,omitempty"`
	Perrla      bool    `gorm:"column:perrla;not null;default:false"        json:"perrla"`
	PerrlaText  *string `gorm:"column:perrla_text;type:varchar(255)"        json:"perrla_text,omitempty"`
	Apd         bool    `gorm:"column:apd;not null;default:false"           json:"apd"`
	ApdText     *string `gorm:"column:apd_text;type:varchar(255)"           json:"apd_text,omitempty"`
}
func (Pupils) TableName() string { return "pupils" }
func (p *Pupils) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_pupils": p.IDPupils, "od_mm_dim": p.OdMmDim, "od_mm_bright": p.OdMmBright,
		"os_mm_dim": p.OsMmDim, "os_mm_bright": p.OsMmBright,
		"perrla": p.Perrla, "perrla_text": p.PerrlaText,
		"apd": p.Apd, "apd_text": p.ApdText,
	}
}

// Bruckner ↔ table: bruckner
type Bruckner struct {
	IDBruckner int64   `gorm:"column:id_bruckner;primaryKey;autoIncrement" json:"id_bruckner"`
	Od         *string `gorm:"column:od;type:varchar(255)"                 json:"od,omitempty"`
	Os         *string `gorm:"column:os;type:varchar(255)"                 json:"os,omitempty"`
	GoodReflex *bool   `gorm:"column:good_reflex"                          json:"good_reflex,omitempty"`
}
func (Bruckner) TableName() string { return "bruckner" }
func (b *Bruckner) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_bruckner": b.IDBruckner, "od": b.Od, "os": b.Os, "good_reflex": b.GoodReflex}
}

// AmslerGrid ↔ table: amsler_grid
type AmslerGrid struct {
	IDAmslerGrid int64   `gorm:"column:id_amsler_grid;primaryKey;autoIncrement" json:"id_amsler_grid"`
	Od           *string `gorm:"column:od;type:varchar(255)"                    json:"od,omitempty"`
	Os           *string `gorm:"column:os;type:varchar(255)"                    json:"os,omitempty"`
}
func (AmslerGrid) TableName() string { return "amsler_grid" }
func (a *AmslerGrid) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_amsler_grid": a.IDAmslerGrid, "od": a.Od, "os": a.Os}
}

// ColorVision ↔ table: color_vision
type ColorVision struct {
	IDColorVision int64   `gorm:"column:id_color_vision;primaryKey;autoIncrement" json:"id_color_vision"`
	Od1           *string `gorm:"column:od1;type:varchar(255)"                    json:"od1,omitempty"`
	Od2           *string `gorm:"column:od2;type:varchar(255)"                    json:"od2,omitempty"`
	Os1           *string `gorm:"column:os1;type:varchar(255)"                    json:"os1,omitempty"`
	Os2           *string `gorm:"column:os2;type:varchar(255)"                    json:"os2,omitempty"`
}
func (ColorVision) TableName() string { return "color_vision" }
func (c *ColorVision) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_color_vision": c.IDColorVision, "od1": c.Od1, "od2": c.Od2, "os1": c.Os1, "os2": c.Os2}
}

// DistanceVonGraefePhoria ↔ table: distance_von_graefe_phoria
type DistanceVonGraefePhoria struct {
	IDDistanceVonGraefePhoria int64   `gorm:"column:id_distance_von_graefe_phoria;primaryKey;autoIncrement" json:"id_distance_von_graefe_phoria"`
	HDistVgp                  *string `gorm:"column:h_dist_vgp;type:varchar(255)"                           json:"h_dist_vgp,omitempty"`
	VDistVgp                  *string `gorm:"column:v_dist_vgp;type:varchar(255)"                           json:"v_dist_vgp,omitempty"`
}
func (DistanceVonGraefePhoria) TableName() string { return "distance_von_graefe_phoria" }
func (d *DistanceVonGraefePhoria) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_distance_von_graefe_phoria": d.IDDistanceVonGraefePhoria, "h_dist_vgp": d.HDistVgp, "v_dist_vgp": d.VDistVgp}
}

// NearVonGraefePhoria ↔ table: near_von_graefe_phoria
type NearVonGraefePhoria struct {
	IDNearVonGraefePhoria int64   `gorm:"column:id_near_von_graefe_phoria;primaryKey;autoIncrement" json:"id_near_von_graefe_phoria"`
	HNearVgp              *string `gorm:"column:h_near_vgp;type:varchar(255)"                       json:"h_near_vgp,omitempty"`
	VNearVgp              *string `gorm:"column:v_near_vgp;type:varchar(255)"                       json:"v_near_vgp,omitempty"`
}
func (NearVonGraefePhoria) TableName() string { return "near_von_graefe_phoria" }
func (n *NearVonGraefePhoria) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_near_von_graefe_phoria": n.IDNearVonGraefePhoria, "h_near_vgp": n.HNearVgp, "v_near_vgp": n.VNearVgp}
}

// BloodPressure ↔ table: blood_pressure
type BloodPressure struct {
	IDBloodPressure int64   `gorm:"column:id_blood_pressure;primaryKey;autoIncrement" json:"id_blood_pressure"`
	Sbp             *string `gorm:"column:sbp;type:varchar(255)"                      json:"sbp,omitempty"`
	Dbp             *string `gorm:"column:dbp;type:varchar(255)"                      json:"dbp,omitempty"`
}
func (BloodPressure) TableName() string { return "blood_pressure" }
func (b *BloodPressure) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_blood_pressure": b.IDBloodPressure, "sbp": b.Sbp, "dbp": b.Dbp}
}

// AutorefractorPreliminary ↔ table: autorefractor_preliminary
type AutorefractorPreliminary struct {
	IDAutorefractorPreliminary int64   `gorm:"column:id_autorefractor_preliminary;primaryKey;autoIncrement" json:"id_autorefractor_preliminary"`
	OdSph                     *string `gorm:"column:od_sph;type:varchar(255)"                              json:"od_sph,omitempty"`
	OsSph                     *string `gorm:"column:os_sph;type:varchar(255)"                              json:"os_sph,omitempty"`
	OdCyl                     *string `gorm:"column:od_cyl;type:varchar(255)"                              json:"od_cyl,omitempty"`
	OsCyl                     *string `gorm:"column:os_cyl;type:varchar(255)"                              json:"os_cyl,omitempty"`
	OdAxis                    *string `gorm:"column:od_axis;type:varchar(255)"                             json:"od_axis,omitempty"`
	OsAxis                    *string `gorm:"column:os_axis;type:varchar(255)"                             json:"os_axis,omitempty"`
	Pd                        *string `gorm:"column:pd;type:varchar(255)"                                  json:"pd,omitempty"`
}
func (AutorefractorPreliminary) TableName() string { return "autorefractor_preliminary" }
func (a *AutorefractorPreliminary) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_autorefractor_preliminary": a.IDAutorefractorPreliminary,
		"od_sph": a.OdSph, "os_sph": a.OsSph, "od_cyl": a.OdCyl, "os_cyl": a.OsCyl,
		"od_axis": a.OdAxis, "os_axis": a.OsAxis, "pd": a.Pd,
	}
}

// AutoKeratometerPreliminary ↔ table: auto_keratometer_preliminary
type AutoKeratometerPreliminary struct {
	IDAutoKeratometerPreliminary int64   `gorm:"column:id_auto_keratometer_preliminary;primaryKey;autoIncrement" json:"id_auto_keratometer_preliminary"`
	OdPw1                        *string `gorm:"column:od_pw1;type:varchar(255)"                                 json:"od_pw1,omitempty"`
	OsPw1                        *string `gorm:"column:os_pw1;type:varchar(255)"                                 json:"os_pw1,omitempty"`
	OdPw2                        *string `gorm:"column:od_pw2;type:varchar(255)"                                 json:"od_pw2,omitempty"`
	OsPw2                        *string `gorm:"column:os_pw2;type:varchar(255)"                                 json:"os_pw2,omitempty"`
	OdAxis                       *string `gorm:"column:od_axis;type:varchar(255)"                                json:"od_axis,omitempty"`
	OsAxis                       *string `gorm:"column:os_axis;type:varchar(255)"                                json:"os_axis,omitempty"`
}
func (AutoKeratometerPreliminary) TableName() string { return "auto_keratometer_preliminary" }
func (a *AutoKeratometerPreliminary) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_auto_keratometer_preliminary": a.IDAutoKeratometerPreliminary,
		"od_pw1": a.OdPw1, "os_pw1": a.OsPw1, "od_pw2": a.OdPw2, "os_pw2": a.OsPw2,
		"od_axis": a.OdAxis, "os_axis": a.OsAxis,
	}
}

// DistPhoriaTest ↔ table: dist_phoria_testing
type DistPhoriaTest struct {
	IDDistPhoriaTest int   `gorm:"column:id_dist_phoria_testing;primaryKey;autoIncrement" json:"id_dist_phoria_testing"`
	Horiz            *string `gorm:"column:horiz;type:varchar(50)"                        json:"horiz,omitempty"`
	Vert             *string `gorm:"column:vert;type:varchar(50)"                         json:"vert,omitempty"`
	HorizExo         bool    `gorm:"column:horiz_exo;not null;default:false"              json:"horiz_exo"`
	HorizEso         bool    `gorm:"column:horiz_eso;not null;default:false"              json:"horiz_eso"`
	HorizOrtho       bool    `gorm:"column:horiz_ortho;not null;default:false"            json:"horiz_ortho"`
	VertRh           bool    `gorm:"column:vert_rh;not null;default:false"                json:"vert_rh"`
	VertLn           bool    `gorm:"column:vert_ln;not null;default:false"                json:"vert_ln"`
	VertOrtho        bool    `gorm:"column:vert_ortho;not null;default:false"             json:"vert_ortho"`
}
func (DistPhoriaTest) TableName() string { return "dist_phoria_testing" }
func (d *DistPhoriaTest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_dist_phoria_testing": d.IDDistPhoriaTest,
		"horiz": d.Horiz, "vert": d.Vert,
		"horiz_exo": d.HorizExo, "horiz_eso": d.HorizEso, "horiz_ortho": d.HorizOrtho,
		"vert_rh": d.VertRh, "vert_ln": d.VertLn, "vert_ortho": d.VertOrtho,
	}
}

// NearPhoriaTest ↔ table: near_phoria_testing
type NearPhoriaTest struct {
	IDNearPhoriaTest  int     `gorm:"column:id_near_phoria_testing;primaryKey;autoIncrement" json:"id_near_phoria_testing"`
	Horiz             *string `gorm:"column:horiz;type:varchar(50)"                          json:"horiz,omitempty"`
	Vert              *string `gorm:"column:vert;type:varchar(50)"                           json:"vert,omitempty"`
	GradientRatio1    *string `gorm:"column:gradient_ratio1;type:varchar(50)"                json:"gradient_ratio1,omitempty"`
	CalculatedRatio1  *string `gorm:"column:calculated_ratio1;type:varchar(50)"              json:"calculated_ratio1,omitempty"`
	GradientRatio2    *string `gorm:"column:gradient_ratio2;type:varchar(50)"                json:"gradient_ratio2,omitempty"`
	CalculatedRatio2  *string `gorm:"column:calculated_ratio2;type:varchar(50)"              json:"calculated_ratio2,omitempty"`
	HorizExo          bool    `gorm:"column:horiz_exo;not null;default:false"                json:"horiz_exo"`
	HorizEso          bool    `gorm:"column:horiz_eso;not null;default:false"                json:"horiz_eso"`
	HorizOrtho        bool    `gorm:"column:horiz_ortho;not null;default:false"              json:"horiz_ortho"`
	VertRh            bool    `gorm:"column:vert_rh;not null;default:false"                  json:"vert_rh"`
	VertLn            bool    `gorm:"column:vert_ln;not null;default:false"                  json:"vert_ln"`
	VertOrtho         bool    `gorm:"column:vert_ortho;not null;default:false"               json:"vert_ortho"`
}
func (NearPhoriaTest) TableName() string { return "near_phoria_testing" }
func (n *NearPhoriaTest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_near_phoria_testing": n.IDNearPhoriaTest,
		"horiz": n.Horiz, "vert": n.Vert,
		"gradient_ratio1": n.GradientRatio1, "calculated_ratio1": n.CalculatedRatio1,
		"gradient_ratio2": n.GradientRatio2, "calculated_ratio2": n.CalculatedRatio2,
		"horiz_exo": n.HorizExo, "horiz_eso": n.HorizEso, "horiz_ortho": n.HorizOrtho,
		"vert_rh": n.VertRh, "vert_ln": n.VertLn, "vert_ortho": n.VertOrtho,
	}
}

// DistVergenceTest ↔ table: dist_vergence_testing
type DistVergenceTest struct {
	IDDistVergenceTest int     `gorm:"column:id_dist_vergence_testing;primaryKey;autoIncrement" json:"id_dist_vergence_testing"`
	Bi1                *string `gorm:"column:bi1;type:varchar(50)"                              json:"bi1,omitempty"`
	Bo1                *string `gorm:"column:bo1;type:varchar(50)"                              json:"bo1,omitempty"`
	Bi2                *string `gorm:"column:bi2;type:varchar(50)"                              json:"bi2,omitempty"`
	Bo2                *string `gorm:"column:bo2;type:varchar(50)"                              json:"bo2,omitempty"`
	Bi3                *string `gorm:"column:bi3;type:varchar(50)"                              json:"bi3,omitempty"`
	Bo3                *string `gorm:"column:bo3;type:varchar(50)"                              json:"bo3,omitempty"`
}
func (DistVergenceTest) TableName() string { return "dist_vergence_testing" }
func (d *DistVergenceTest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_dist_vergence_testing": d.IDDistVergenceTest,
		"bi1": d.Bi1, "bo1": d.Bo1, "bi2": d.Bi2, "bo2": d.Bo2, "bi3": d.Bi3, "bo3": d.Bo3,
	}
}

// NearVergenceTest ↔ table: near_vergence_testing
type NearVergenceTest struct {
	IDNearVergenceTest int     `gorm:"column:id_near_vergence_testing;primaryKey;autoIncrement" json:"id_near_vergence_testing"`
	Bi1                *string `gorm:"column:bi1;type:varchar(50)"                              json:"bi1,omitempty"`
	Bo1                *string `gorm:"column:bo1;type:varchar(50)"                              json:"bo1,omitempty"`
	Bi2                *string `gorm:"column:bi2;type:varchar(50)"                              json:"bi2,omitempty"`
	Bo2                *string `gorm:"column:bo2;type:varchar(50)"                              json:"bo2,omitempty"`
	Bi3                *string `gorm:"column:bi3;type:varchar(50)"                              json:"bi3,omitempty"`
	Bo3                *string `gorm:"column:bo3;type:varchar(50)"                              json:"bo3,omitempty"`
}
func (NearVergenceTest) TableName() string { return "near_vergence_testing" }
func (n *NearVergenceTest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_near_vergence_testing": n.IDNearVergenceTest,
		"bi1": n.Bi1, "bo1": n.Bo1, "bi2": n.Bi2, "bo2": n.Bo2, "bi3": n.Bi3, "bo3": n.Bo3,
	}
}

// Accommodation ↔ table: accommodation
type Accommodation struct {
	IDAccommodation       int     `gorm:"column:id_accommodation;primaryKey;autoIncrement" json:"id_accommodation"`
	Pra1                  *string `gorm:"column:pra1;type:varchar(50)"                     json:"pra1,omitempty"`
	Nra1                  *string `gorm:"column:nra1;type:varchar(50)"                     json:"nra1,omitempty"`
	Pra2                  *string `gorm:"column:pra2;type:varchar(50)"                     json:"pra2,omitempty"`
	Nra2                  *string `gorm:"column:nra2;type:varchar(50)"                     json:"nra2,omitempty"`
	MemOd                 *string `gorm:"column:mem_od;type:varchar(50)"                   json:"mem_od,omitempty"`
	MemOs                 *string `gorm:"column:mem_os;type:varchar(50)"                   json:"mem_os,omitempty"`
	Baf                   *string `gorm:"column:baf;type:varchar(50)"                      json:"baf,omitempty"`
	VergenceFacilityCpm   *string `gorm:"column:vergence_facility_cpm;type:varchar(50)"    json:"vergence_facility_cpm,omitempty"`
	VergenceFacilityWith  *string `gorm:"column:vergence_facility_with;type:varchar(50)"   json:"vergence_facility_with,omitempty"`
	PushUpOd              *string `gorm:"column:push_up_od;type:varchar(50)"               json:"push_up_od,omitempty"`
	PushUpOs              *string `gorm:"column:push_up_os;type:varchar(50)"               json:"push_up_os,omitempty"`
	PushUpOu              *string `gorm:"column:push_up_ou;type:varchar(50)"               json:"push_up_ou,omitempty"`
	SlowWith              *bool   `gorm:"column:slow_with"                                 json:"slow_with,omitempty"`
}
func (Accommodation) TableName() string { return "accommodation" }
func (a *Accommodation) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_accommodation": a.IDAccommodation,
		"pra1": a.Pra1, "nra1": a.Nra1, "pra2": a.Pra2, "nra2": a.Nra2,
		"mem_od": a.MemOd, "mem_os": a.MemOs, "baf": a.Baf,
		"vergence_facility_cpm": a.VergenceFacilityCpm, "vergence_facility_with": a.VergenceFacilityWith,
		"push_up_od": a.PushUpOd, "push_up_os": a.PushUpOs, "push_up_ou": a.PushUpOu,
		"slow_with": a.SlowWith,
	}
}

// NearPointTesting ↔ table: near_point_testing
type NearPointTesting struct {
	IDNearPointTesting    int64  `gorm:"column:id_near_point_testing;primaryKey;autoIncrement" json:"id_near_point_testing"`
	DistPhoriaTestingID   *int64 `gorm:"column:dist_phoria_testing_id"                         json:"dist_phoria_testing_id,omitempty"`
	NearPhoriaTestingID   *int64 `gorm:"column:near_phoria_testing_id"                         json:"near_phoria_testing_id,omitempty"`
	DistVergenceTestingID *int64 `gorm:"column:dist_vergence_testing_id"                       json:"dist_vergence_testing_id,omitempty"`
	NearVergenceTestingID *int64 `gorm:"column:near_vergence_testing_id"                       json:"near_vergence_testing_id,omitempty"`
	AccommodationID       *int64 `gorm:"column:accommodation_id"                               json:"accommodation_id,omitempty"`
}
func (NearPointTesting) TableName() string { return "near_point_testing" }
func (n *NearPointTesting) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_near_point_testing":    n.IDNearPointTesting,
		"dist_phoria_testing_id":   n.DistPhoriaTestingID,
		"near_phoria_testing_id":   n.NearPhoriaTestingID,
		"dist_vergence_testing_id": n.DistVergenceTestingID,
		"near_vergence_testing_id": n.NearVergenceTestingID,
		"accommodation_id":         n.AccommodationID,
	}
}
