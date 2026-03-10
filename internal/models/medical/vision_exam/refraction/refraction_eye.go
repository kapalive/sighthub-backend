package refraction

// RefractionEye ↔ table: refraction_eye
type RefractionEye struct {
	IDRefractionEye int64   `gorm:"column:id_refraction_eye;primaryKey;autoIncrement" json:"id_refraction_eye"`
	RetinoscopyID   int64   `gorm:"column:retinoscopy_id;not null;uniqueIndex"         json:"retinoscopy_id"`
	CycloID         int64   `gorm:"column:cyclo_id;not null;uniqueIndex"               json:"cyclo_id"`
	ManifestID      int64   `gorm:"column:manifest_id;not null;uniqueIndex"            json:"manifest_id"`
	FinalID         int64   `gorm:"column:final_id;not null;uniqueIndex"               json:"final_id"`
	Final2ID        *int64  `gorm:"column:final2_id"                                   json:"final2_id,omitempty"`
	Final3ID        *int64  `gorm:"column:final3_id"                                   json:"final3_id,omitempty"`
	DrNote          *string `gorm:"column:dr_note;type:text"                           json:"dr_note,omitempty"`
	EyeExamID       int64   `gorm:"column:eye_exam_id;not null"                        json:"eye_exam_id"`

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
