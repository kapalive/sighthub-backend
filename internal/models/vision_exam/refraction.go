// internal/models/vision_exam/refraction.go
// Refraction sub-tables and main refraction_eye container.
package vision_exam

import "time"

// refrRow is the common set of fields shared by Retinoscopy/Cyclo/Manifest/Final/Final2/Final3.
// We embed it in each struct to avoid repetition.

// Retinoscopy ↔ table: retinoscopy
type Retinoscopy struct {
	IDRetinoscopy   int64   `gorm:"column:id_retinoscopy;primaryKey;autoIncrement" json:"id_retinoscopy"`
	OdSph           *string `gorm:"column:od_sph;type:varchar(255)"  json:"od_sph,omitempty"`
	OsSph           *string `gorm:"column:os_sph;type:varchar(255)"  json:"os_sph,omitempty"`
	OdCyl           *string `gorm:"column:od_cyl;type:varchar(255)"  json:"od_cyl,omitempty"`
	OsCyl           *string `gorm:"column:os_cyl;type:varchar(255)"  json:"os_cyl,omitempty"`
	OdAxis          *string `gorm:"column:od_axis;type:varchar(255)" json:"od_axis,omitempty"`
	OsAxis          *string `gorm:"column:os_axis;type:varchar(255)" json:"os_axis,omitempty"`
	OdAdd           *string `gorm:"column:od_add;type:varchar(255)"  json:"od_add,omitempty"`
	OsAdd           *string `gorm:"column:os_add;type:varchar(255)"  json:"os_add,omitempty"`
	OdHPrism        *string `gorm:"column:od_h_prism;type:varchar(255)"      json:"od_h_prism,omitempty"`
	OdHPrismList    *string `gorm:"column:od_h_prism_list;type:varchar(255)" json:"od_h_prism_list,omitempty"`
	OsHPrism        *string `gorm:"column:os_h_prism;type:varchar(255)"      json:"os_h_prism,omitempty"`
	OsHPrismList    *string `gorm:"column:os_h_prism_list;type:varchar(255)" json:"os_h_prism_list,omitempty"`
	OdVPrism        *string `gorm:"column:od_v_prism;type:varchar(255)"      json:"od_v_prism,omitempty"`
	OdVPrismList    *string `gorm:"column:od_v_prism_list;type:varchar(255)" json:"od_v_prism_list,omitempty"`
	OsVPrism        *string `gorm:"column:os_v_prism;type:varchar(255)"      json:"os_v_prism,omitempty"`
	OsVPrismList    *string `gorm:"column:os_v_prism_list;type:varchar(255)" json:"os_v_prism_list,omitempty"`
	OdDva20         *string `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20         *string `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OdNva20         *string `gorm:"column:od_nva_20;type:varchar(255)" json:"od_nva_20,omitempty"`
	OsNva20         *string `gorm:"column:os_nva_20;type:varchar(255)" json:"os_nva_20,omitempty"`
	OuDva20         *string `gorm:"column:ou_dva_20;type:varchar(255)" json:"ou_dva_20,omitempty"`
	OuNva20         *string `gorm:"column:ou_nva_20;type:varchar(255)" json:"ou_nva_20,omitempty"`
	OdFinal         bool    `gorm:"column:od_final;not null;default:false" json:"od_final"`
	OsFinal         bool    `gorm:"column:os_final;not null;default:false" json:"os_final"`
}
func (Retinoscopy) TableName() string { return "retinoscopy" }
func (r *Retinoscopy) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_retinoscopy": r.IDRetinoscopy,
		"od_sph": r.OdSph, "os_sph": r.OsSph, "od_cyl": r.OdCyl, "os_cyl": r.OsCyl,
		"od_axis": r.OdAxis, "os_axis": r.OsAxis, "od_add": r.OdAdd, "os_add": r.OsAdd,
		"od_h_prism": r.OdHPrism, "od_h_prism_list": r.OdHPrismList,
		"os_h_prism": r.OsHPrism, "os_h_prism_list": r.OsHPrismList,
		"od_v_prism": r.OdVPrism, "od_v_prism_list": r.OdVPrismList,
		"os_v_prism": r.OsVPrism, "os_v_prism_list": r.OsVPrismList,
		"od_dva_20": r.OdDva20, "os_dva_20": r.OsDva20, "od_nva_20": r.OdNva20,
		"os_nva_20": r.OsNva20, "ou_dva_20": r.OuDva20, "ou_nva_20": r.OuNva20,
		"od_final": r.OdFinal, "os_final": r.OsFinal,
	}
}

// Cyclo ↔ table: cyclo (cycloplegic refraction)
type Cyclo struct {
	IDCyclo         int64   `gorm:"column:id_cyclo;primaryKey;autoIncrement" json:"id_cyclo"`
	OdSph           *string `gorm:"column:od_sph;type:varchar(255)"  json:"od_sph,omitempty"`
	OsSph           *string `gorm:"column:os_sph;type:varchar(255)"  json:"os_sph,omitempty"`
	OdCyl           *string `gorm:"column:od_cyl;type:varchar(255)"  json:"od_cyl,omitempty"`
	OsCyl           *string `gorm:"column:os_cyl;type:varchar(255)"  json:"os_cyl,omitempty"`
	OdAxis          *string `gorm:"column:od_axis;type:varchar(255)" json:"od_axis,omitempty"`
	OsAxis          *string `gorm:"column:os_axis;type:varchar(255)" json:"os_axis,omitempty"`
	OdAdd           *string `gorm:"column:od_add;type:varchar(255)"  json:"od_add,omitempty"`
	OsAdd           *string `gorm:"column:os_add;type:varchar(255)"  json:"os_add,omitempty"`
	OdHPrism        *string `gorm:"column:od_h_prism;type:varchar(255)"      json:"od_h_prism,omitempty"`
	OdHPrismList    *string `gorm:"column:od_h_prism_list;type:varchar(255)" json:"od_h_prism_list,omitempty"`
	OsHPrism        *string `gorm:"column:os_h_prism;type:varchar(255)"      json:"os_h_prism,omitempty"`
	OsHPrismList    *string `gorm:"column:os_h_prism_list;type:varchar(255)" json:"os_h_prism_list,omitempty"`
	OdVPrism        *string `gorm:"column:od_v_prism;type:varchar(255)"      json:"od_v_prism,omitempty"`
	OdVPrismList    *string `gorm:"column:od_v_prism_list;type:varchar(255)" json:"od_v_prism_list,omitempty"`
	OsVPrism        *string `gorm:"column:os_v_prism;type:varchar(255)"      json:"os_v_prism,omitempty"`
	OsVPrismList    *string `gorm:"column:os_v_prism_list;type:varchar(255)" json:"os_v_prism_list,omitempty"`
	OdDva20         *string `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20         *string `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OdNva20         *string `gorm:"column:od_nva_20;type:varchar(255)" json:"od_nva_20,omitempty"`
	OsNva20         *string `gorm:"column:os_nva_20;type:varchar(255)" json:"os_nva_20,omitempty"`
	OuDva20         *string `gorm:"column:ou_dva_20;type:varchar(255)" json:"ou_dva_20,omitempty"`
	OuNva20         *string `gorm:"column:ou_nva_20;type:varchar(255)" json:"ou_nva_20,omitempty"`
	OdFinal         bool    `gorm:"column:od_final;not null;default:false" json:"od_final"`
	OsFinal         bool    `gorm:"column:os_final;not null;default:false" json:"os_final"`
	OdPd            *string `gorm:"column:od_pd;type:varchar(255)"  json:"od_pd,omitempty"`
	OsPd            *string `gorm:"column:os_pd;type:varchar(255)"  json:"os_pd,omitempty"`
	OuPd            *string `gorm:"column:ou_pd;type:varchar(255)"  json:"ou_pd,omitempty"`
	OdNpd           *string `gorm:"column:od_npd;type:varchar(255)" json:"od_npd,omitempty"`
	OsNpd           *string `gorm:"column:os_npd;type:varchar(255)" json:"os_npd,omitempty"`
	OuNpd           *string `gorm:"column:ou_npd;type:varchar(255)" json:"ou_npd,omitempty"`
}
func (Cyclo) TableName() string { return "cyclo" }
func (c *Cyclo) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_cyclo": c.IDCyclo,
		"od_sph": c.OdSph, "os_sph": c.OsSph, "od_cyl": c.OdCyl, "os_cyl": c.OsCyl,
		"od_axis": c.OdAxis, "os_axis": c.OsAxis, "od_add": c.OdAdd, "os_add": c.OsAdd,
		"od_h_prism": c.OdHPrism, "od_h_prism_list": c.OdHPrismList,
		"os_h_prism": c.OsHPrism, "os_h_prism_list": c.OsHPrismList,
		"od_v_prism": c.OdVPrism, "od_v_prism_list": c.OdVPrismList,
		"os_v_prism": c.OsVPrism, "os_v_prism_list": c.OsVPrismList,
		"od_dva_20": c.OdDva20, "os_dva_20": c.OsDva20, "od_nva_20": c.OdNva20,
		"os_nva_20": c.OsNva20, "ou_dva_20": c.OuDva20, "ou_nva_20": c.OuNva20,
		"od_final": c.OdFinal, "os_final": c.OsFinal,
		"od_pd": c.OdPd, "os_pd": c.OsPd, "ou_pd": c.OuPd,
		"od_npd": c.OdNpd, "os_npd": c.OsNpd, "ou_npd": c.OuNpd,
	}
}

// Manifest ↔ table: manifest (same fields as Cyclo)
type Manifest struct {
	IDManifest      int64   `gorm:"column:id_manifest;primaryKey;autoIncrement" json:"id_manifest"`
	OdSph           *string `gorm:"column:od_sph;type:varchar(255)"  json:"od_sph,omitempty"`
	OsSph           *string `gorm:"column:os_sph;type:varchar(255)"  json:"os_sph,omitempty"`
	OdCyl           *string `gorm:"column:od_cyl;type:varchar(255)"  json:"od_cyl,omitempty"`
	OsCyl           *string `gorm:"column:os_cyl;type:varchar(255)"  json:"os_cyl,omitempty"`
	OdAxis          *string `gorm:"column:od_axis;type:varchar(255)" json:"od_axis,omitempty"`
	OsAxis          *string `gorm:"column:os_axis;type:varchar(255)" json:"os_axis,omitempty"`
	OdAdd           *string `gorm:"column:od_add;type:varchar(255)"  json:"od_add,omitempty"`
	OsAdd           *string `gorm:"column:os_add;type:varchar(255)"  json:"os_add,omitempty"`
	OdHPrism        *string `gorm:"column:od_h_prism;type:varchar(255)"      json:"od_h_prism,omitempty"`
	OdHPrismList    *string `gorm:"column:od_h_prism_list;type:varchar(255)" json:"od_h_prism_list,omitempty"`
	OsHPrism        *string `gorm:"column:os_h_prism;type:varchar(255)"      json:"os_h_prism,omitempty"`
	OsHPrismList    *string `gorm:"column:os_h_prism_list;type:varchar(255)" json:"os_h_prism_list,omitempty"`
	OdVPrism        *string `gorm:"column:od_v_prism;type:varchar(255)"      json:"od_v_prism,omitempty"`
	OdVPrismList    *string `gorm:"column:od_v_prism_list;type:varchar(255)" json:"od_v_prism_list,omitempty"`
	OsVPrism        *string `gorm:"column:os_v_prism;type:varchar(255)"      json:"os_v_prism,omitempty"`
	OsVPrismList    *string `gorm:"column:os_v_prism_list;type:varchar(255)" json:"os_v_prism_list,omitempty"`
	OdDva20         *string `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20         *string `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OdNva20         *string `gorm:"column:od_nva_20;type:varchar(255)" json:"od_nva_20,omitempty"`
	OsNva20         *string `gorm:"column:os_nva_20;type:varchar(255)" json:"os_nva_20,omitempty"`
	OuDva20         *string `gorm:"column:ou_dva_20;type:varchar(255)" json:"ou_dva_20,omitempty"`
	OuNva20         *string `gorm:"column:ou_nva_20;type:varchar(255)" json:"ou_nva_20,omitempty"`
	OdFinal         bool    `gorm:"column:od_final;not null;default:false" json:"od_final"`
	OsFinal         bool    `gorm:"column:os_final;not null;default:false" json:"os_final"`
	OdPd            *string `gorm:"column:od_pd;type:varchar(255)"  json:"od_pd,omitempty"`
	OsPd            *string `gorm:"column:os_pd;type:varchar(255)"  json:"os_pd,omitempty"`
	OuPd            *string `gorm:"column:ou_pd;type:varchar(255)"  json:"ou_pd,omitempty"`
	OdNpd           *string `gorm:"column:od_npd;type:varchar(255)" json:"od_npd,omitempty"`
	OsNpd           *string `gorm:"column:os_npd;type:varchar(255)" json:"os_npd,omitempty"`
	OuNpd           *string `gorm:"column:ou_npd;type:varchar(255)" json:"ou_npd,omitempty"`
}
func (Manifest) TableName() string { return "manifest" }
func (m *Manifest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_manifest": m.IDManifest,
		"od_sph": m.OdSph, "os_sph": m.OsSph, "od_cyl": m.OdCyl, "os_cyl": m.OsCyl,
		"od_axis": m.OdAxis, "os_axis": m.OsAxis, "od_add": m.OdAdd, "os_add": m.OsAdd,
		"od_h_prism": m.OdHPrism, "od_h_prism_list": m.OdHPrismList,
		"os_h_prism": m.OsHPrism, "os_h_prism_list": m.OsHPrismList,
		"od_v_prism": m.OdVPrism, "od_v_prism_list": m.OdVPrismList,
		"os_v_prism": m.OsVPrism, "os_v_prism_list": m.OsVPrismList,
		"od_dva_20": m.OdDva20, "os_dva_20": m.OsDva20, "od_nva_20": m.OdNva20,
		"os_nva_20": m.OsNva20, "ou_dva_20": m.OuDva20, "ou_nva_20": m.OuNva20,
		"od_final": m.OdFinal, "os_final": m.OsFinal,
		"od_pd": m.OdPd, "os_pd": m.OsPd, "ou_pd": m.OuPd,
		"od_npd": m.OdNpd, "os_npd": m.OsNpd, "ou_npd": m.OuNpd,
	}
}

// RefractionFinal ↔ table: final
type RefractionFinal struct {
	IDFinal         int64      `gorm:"column:id_final;primaryKey;autoIncrement" json:"id_final"`
	OdSph           *string    `gorm:"column:od_sph;type:varchar(255)"  json:"od_sph,omitempty"`
	OsSph           *string    `gorm:"column:os_sph;type:varchar(255)"  json:"os_sph,omitempty"`
	OdCyl           *string    `gorm:"column:od_cyl;type:varchar(255)"  json:"od_cyl,omitempty"`
	OsCyl           *string    `gorm:"column:os_cyl;type:varchar(255)"  json:"os_cyl,omitempty"`
	OdAxis          *string    `gorm:"column:od_axis;type:varchar(255)" json:"od_axis,omitempty"`
	OsAxis          *string    `gorm:"column:os_axis;type:varchar(255)" json:"os_axis,omitempty"`
	OdAdd           *string    `gorm:"column:od_add;type:varchar(255)"  json:"od_add,omitempty"`
	OsAdd           *string    `gorm:"column:os_add;type:varchar(255)"  json:"os_add,omitempty"`
	OdHPrism        *string    `gorm:"column:od_h_prism;type:varchar(255)"      json:"od_h_prism,omitempty"`
	OdHPrismList    *string    `gorm:"column:od_h_prism_list;type:varchar(255)" json:"od_h_prism_list,omitempty"`
	OsHPrism        *string    `gorm:"column:os_h_prism;type:varchar(255)"      json:"os_h_prism,omitempty"`
	OsHPrismList    *string    `gorm:"column:os_h_prism_list;type:varchar(255)" json:"os_h_prism_list,omitempty"`
	OdVPrism        *string    `gorm:"column:od_v_prism;type:varchar(255)"      json:"od_v_prism,omitempty"`
	OdVPrismList    *string    `gorm:"column:od_v_prism_list;type:varchar(255)" json:"od_v_prism_list,omitempty"`
	OsVPrism        *string    `gorm:"column:os_v_prism;type:varchar(255)"      json:"os_v_prism,omitempty"`
	OsVPrismList    *string    `gorm:"column:os_v_prism_list;type:varchar(255)" json:"os_v_prism_list,omitempty"`
	OdDva20         *string    `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20         *string    `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OdNva20         *string    `gorm:"column:od_nva_20;type:varchar(255)" json:"od_nva_20,omitempty"`
	OsNva20         *string    `gorm:"column:os_nva_20;type:varchar(255)" json:"os_nva_20,omitempty"`
	OuDva20         *string    `gorm:"column:ou_dva_20;type:varchar(255)" json:"ou_dva_20,omitempty"`
	OuNva20         *string    `gorm:"column:ou_nva_20;type:varchar(255)" json:"ou_nva_20,omitempty"`
	OdPd            *string    `gorm:"column:od_pd;type:varchar(255)"  json:"od_pd,omitempty"`
	OsPd            *string    `gorm:"column:os_pd;type:varchar(255)"  json:"os_pd,omitempty"`
	OuPd            *string    `gorm:"column:ou_pd;type:varchar(255)"  json:"ou_pd,omitempty"`
	OdNpd           *string    `gorm:"column:od_npd;type:varchar(255)" json:"od_npd,omitempty"`
	OsNpd           *string    `gorm:"column:os_npd;type:varchar(255)" json:"os_npd,omitempty"`
	OuNpd           *string    `gorm:"column:ou_npd;type:varchar(255)" json:"ou_npd,omitempty"`
	ExpireDate      *time.Time `gorm:"column:expire_date;type:date"    json:"expire_date,omitempty"`
	Note            *string    `gorm:"column:note;type:varchar(255)"   json:"note,omitempty"`
}
func (RefractionFinal) TableName() string { return "final" }
func (f *RefractionFinal) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_final": f.IDFinal,
		"od_sph": f.OdSph, "os_sph": f.OsSph, "od_cyl": f.OdCyl, "os_cyl": f.OsCyl,
		"od_axis": f.OdAxis, "os_axis": f.OsAxis, "od_add": f.OdAdd, "os_add": f.OsAdd,
		"od_h_prism": f.OdHPrism, "od_h_prism_list": f.OdHPrismList,
		"os_h_prism": f.OsHPrism, "os_h_prism_list": f.OsHPrismList,
		"od_v_prism": f.OdVPrism, "od_v_prism_list": f.OdVPrismList,
		"os_v_prism": f.OsVPrism, "os_v_prism_list": f.OsVPrismList,
		"od_dva_20": f.OdDva20, "os_dva_20": f.OsDva20, "od_nva_20": f.OdNva20,
		"os_nva_20": f.OsNva20, "ou_dva_20": f.OuDva20, "ou_nva_20": f.OuNva20,
		"od_pd": f.OdPd, "os_pd": f.OsPd, "ou_pd": f.OuPd,
		"od_npd": f.OdNpd, "os_npd": f.OsNpd, "ou_npd": f.OuNpd,
		"note": f.Note,
	}
	if f.ExpireDate != nil {
		m["expire_date"] = f.ExpireDate.Format("2006-01-02")
	} else {
		m["expire_date"] = nil
	}
	return m
}

// Final2 ↔ table: final2 (optional second final Rx)
type Final2 struct {
	IDFinal2        int64   `gorm:"column:id_final;primaryKey;autoIncrement" json:"id_final"`
	OdSph           *string `gorm:"column:od_sph;type:varchar(255)"  json:"od_sph,omitempty"`
	OsSph           *string `gorm:"column:os_sph;type:varchar(255)"  json:"os_sph,omitempty"`
	OdCyl           *string `gorm:"column:od_cyl;type:varchar(255)"  json:"od_cyl,omitempty"`
	OsCyl           *string `gorm:"column:os_cyl;type:varchar(255)"  json:"os_cyl,omitempty"`
	OdAxis          *string `gorm:"column:od_axis;type:varchar(255)" json:"od_axis,omitempty"`
	OsAxis          *string `gorm:"column:os_axis;type:varchar(255)" json:"os_axis,omitempty"`
	OdAdd           *string `gorm:"column:od_add;type:varchar(255)"  json:"od_add,omitempty"`
	OsAdd           *string `gorm:"column:os_add;type:varchar(255)"  json:"os_add,omitempty"`
	OdHPrism        *string `gorm:"column:od_h_prism;type:varchar(255)"      json:"od_h_prism,omitempty"`
	OdHPrismList    *string `gorm:"column:od_h_prism_list;type:varchar(255)" json:"od_h_prism_list,omitempty"`
	OsHPrism        *string `gorm:"column:os_h_prism;type:varchar(255)"      json:"os_h_prism,omitempty"`
	OsHPrismList    *string `gorm:"column:os_h_prism_list;type:varchar(255)" json:"os_h_prism_list,omitempty"`
	OdVPrism        *string `gorm:"column:od_v_prism;type:varchar(255)"      json:"od_v_prism,omitempty"`
	OdVPrismList    *string `gorm:"column:od_v_prism_list;type:varchar(255)" json:"od_v_prism_list,omitempty"`
	OsVPrism        *string `gorm:"column:os_v_prism;type:varchar(255)"      json:"os_v_prism,omitempty"`
	OsVPrismList    *string `gorm:"column:os_v_prism_list;type:varchar(255)" json:"os_v_prism_list,omitempty"`
	OdDva20         *string `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20         *string `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OdNva20         *string `gorm:"column:od_nva_20;type:varchar(255)" json:"od_nva_20,omitempty"`
	OsNva20         *string `gorm:"column:os_nva_20;type:varchar(255)" json:"os_nva_20,omitempty"`
	OuDva20         *string `gorm:"column:ou_dva_20;type:varchar(255)" json:"ou_dva_20,omitempty"`
	OuNva20         *string `gorm:"column:ou_nva_20;type:varchar(255)" json:"ou_nva_20,omitempty"`
	OdPd            *string `gorm:"column:od_pd;type:varchar(255)"  json:"od_pd,omitempty"`
	OsPd            *string `gorm:"column:os_pd;type:varchar(255)"  json:"os_pd,omitempty"`
	OuPd            *string `gorm:"column:ou_pd;type:varchar(255)"  json:"ou_pd,omitempty"`
	OdNpd           *string `gorm:"column:od_npd;type:varchar(255)" json:"od_npd,omitempty"`
	OsNpd           *string `gorm:"column:os_npd;type:varchar(255)" json:"os_npd,omitempty"`
	OuNpd           *string `gorm:"column:ou_npd;type:varchar(255)" json:"ou_npd,omitempty"`
	Desc            *string `gorm:"column:desc;type:varchar(255)"   json:"desc,omitempty"`
	Note            *string `gorm:"column:note;type:varchar(255)"   json:"note,omitempty"`
}
func (Final2) TableName() string { return "final2" }
func (f *Final2) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_final": f.IDFinal2,
		"od_sph": f.OdSph, "os_sph": f.OsSph, "od_cyl": f.OdCyl, "os_cyl": f.OsCyl,
		"od_axis": f.OdAxis, "os_axis": f.OsAxis, "od_add": f.OdAdd, "os_add": f.OsAdd,
		"od_dva_20": f.OdDva20, "os_dva_20": f.OsDva20, "ou_dva_20": f.OuDva20,
		"od_npd": f.OdNpd, "os_npd": f.OsNpd, "ou_npd": f.OuNpd,
		"desc": f.Desc, "note": f.Note,
	}
}

// Final3 ↔ table: final3 (optional third final Rx)
type Final3 struct {
	IDFinal3        int64   `gorm:"column:id_final;primaryKey;autoIncrement" json:"id_final"`
	OdSph           *string `gorm:"column:od_sph;type:varchar(255)"  json:"od_sph,omitempty"`
	OsSph           *string `gorm:"column:os_sph;type:varchar(255)"  json:"os_sph,omitempty"`
	OdCyl           *string `gorm:"column:od_cyl;type:varchar(255)"  json:"od_cyl,omitempty"`
	OsCyl           *string `gorm:"column:os_cyl;type:varchar(255)"  json:"os_cyl,omitempty"`
	OdAxis          *string `gorm:"column:od_axis;type:varchar(255)" json:"od_axis,omitempty"`
	OsAxis          *string `gorm:"column:os_axis;type:varchar(255)" json:"os_axis,omitempty"`
	OdAdd           *string `gorm:"column:od_add;type:varchar(255)"  json:"od_add,omitempty"`
	OsAdd           *string `gorm:"column:os_add;type:varchar(255)"  json:"os_add,omitempty"`
	OdDva20         *string `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20         *string `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OuDva20         *string `gorm:"column:ou_dva_20;type:varchar(255)" json:"ou_dva_20,omitempty"`
	OdPd            *string `gorm:"column:od_pd;type:varchar(255)"  json:"od_pd,omitempty"`
	OsPd            *string `gorm:"column:os_pd;type:varchar(255)"  json:"os_pd,omitempty"`
	OuPd            *string `gorm:"column:ou_pd;type:varchar(255)"  json:"ou_pd,omitempty"`
	OdNpd           *string `gorm:"column:od_npd;type:varchar(255)" json:"od_npd,omitempty"`
	OsNpd           *string `gorm:"column:os_npd;type:varchar(255)" json:"os_npd,omitempty"`
	OuNpd           *string `gorm:"column:ou_npd;type:varchar(255)" json:"ou_npd,omitempty"`
}
func (Final3) TableName() string { return "final3" }
func (f *Final3) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_final": f.IDFinal3,
		"od_sph": f.OdSph, "os_sph": f.OsSph, "od_cyl": f.OdCyl, "os_cyl": f.OsCyl,
		"od_axis": f.OdAxis, "os_axis": f.OsAxis, "od_add": f.OdAdd, "os_add": f.OsAdd,
		"od_dva_20": f.OdDva20, "os_dva_20": f.OsDva20, "ou_dva_20": f.OuDva20,
		"od_npd": f.OdNpd, "os_npd": f.OsNpd, "ou_npd": f.OuNpd,
	}
}

// RefractionEye ↔ table: refraction_eye
type RefractionEye struct {
	IDRefractionEye int64  `gorm:"column:id_refraction_eye;primaryKey;autoIncrement" json:"id_refraction_eye"`
	RetinoscopyID   int64  `gorm:"column:retinoscopy_id;not null;uniqueIndex"         json:"retinoscopy_id"`
	CycloID         int64  `gorm:"column:cyclo_id;not null;uniqueIndex"               json:"cyclo_id"`
	ManifestID      int64  `gorm:"column:manifest_id;not null;uniqueIndex"            json:"manifest_id"`
	FinalID         int64  `gorm:"column:final_id;not null;uniqueIndex"               json:"final_id"`
	Final2ID        *int64 `gorm:"column:final2_id"                                   json:"final2_id,omitempty"`
	Final3ID        *int64 `gorm:"column:final3_id"                                   json:"final3_id,omitempty"`
	DrNote          *string `gorm:"column:dr_note;type:text"                          json:"dr_note,omitempty"`
	EyeExamID       int64  `gorm:"column:eye_exam_id;not null"                        json:"eye_exam_id"`

	Retinoscopy *Retinoscopy    `gorm:"foreignKey:RetinoscopyID;references:IDRetinoscopy" json:"-"`
	Cyclo       *Cyclo          `gorm:"foreignKey:CycloID;references:IDCyclo"             json:"-"`
	Manifest    *Manifest       `gorm:"foreignKey:ManifestID;references:IDManifest"       json:"-"`
	Final       *RefractionFinal `gorm:"foreignKey:FinalID;references:IDFinal"            json:"-"`
	Final2      *Final2         `gorm:"foreignKey:Final2ID;references:IDFinal2"           json:"-"`
	Final3      *Final3         `gorm:"foreignKey:Final3ID;references:IDFinal3"           json:"-"`
}
func (RefractionEye) TableName() string { return "refraction_eye" }
func (r *RefractionEye) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_refraction_eye": r.IDRefractionEye,
		"retinoscopy_id": r.RetinoscopyID, "cyclo_id": r.CycloID,
		"manifest_id": r.ManifestID, "final_id": r.FinalID,
		"final2_id": r.Final2ID, "final3_id": r.Final3ID,
		"dr_note": r.DrNote, "eye_exam_id": r.EyeExamID,
	}
	if r.Retinoscopy != nil { m["retinoscopy"] = r.Retinoscopy.ToMap() }
	if r.Cyclo != nil       { m["cyclo"]       = r.Cyclo.ToMap() }
	if r.Manifest != nil    { m["manifest"]    = r.Manifest.ToMap() }
	if r.Final != nil       { m["final"]       = r.Final.ToMap() }
	if r.Final2 != nil      { m["final2"]      = r.Final2.ToMap() }
	if r.Final3 != nil      { m["final3"]      = r.Final3.ToMap() }
	return m
}
