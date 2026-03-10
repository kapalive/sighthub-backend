package external_sle

type ExternalSleEye struct {
	IDExternalSleEye        int64   `gorm:"column:id_external_sle_eye;primaryKey;autoIncrement" json:"id_external_sle_eye"`
	FindingsExternalSleID   int64   `gorm:"column:findings_external_sle_id;not null;uniqueIndex" json:"findings_external_sle_id"`
	AddDrawing              *string `gorm:"column:add_drawing;type:text" json:"add_drawing"`
	GonioscopyExternalSleID int64   `gorm:"column:gonioscopy_external_sle_id;not null;uniqueIndex" json:"gonioscopy_external_sle_id"`
	PachExternalSleID       int64   `gorm:"column:pach_external_sle_id;not null;uniqueIndex" json:"pach_external_sle_id"`
	OdAngleEstimation       string  `gorm:"column:od_angle_estimation;not null;default:n/a" json:"od_angle_estimation"`
	OsAngleEstimation       string  `gorm:"column:os_angle_estimation;not null;default:n/a" json:"os_angle_estimation"`
	IopDropsFluress         *bool   `gorm:"column:iop_drops_fluress;default:false" json:"iop_drops_fluress"`
	IopDropsProparacaine    *bool   `gorm:"column:iop_drops_proparacaine;default:false" json:"iop_drops_proparacaine"`
	IopDropsFluoroStrip     *bool   `gorm:"column:iop_drops_fluoro_strip;default:false" json:"iop_drops_fluoro_strip"`
	Note                    *string `gorm:"column:note;type:text" json:"note"`
	EyeExamID               int64   `gorm:"column:eye_exam_id;not null" json:"eye_exam_id"`
	VisualFieldsID          *int64  `gorm:"column:visual_fields_id" json:"visual_fields_id"`

	Findings      FindingsExternalSle   `gorm:"foreignKey:IDFindingsExternalSle;references:FindingsExternalSleID" json:"findings"`
	Gonioscopy    GonioscopyExternalSle `gorm:"foreignKey:IDGonioscopyExternalSle;references:GonioscopyExternalSleID" json:"gonioscopy"`
	Pach          PachExternalSle       `gorm:"foreignKey:IDPachExternalSle;references:PachExternalSleID" json:"pach"`
	VisualFields  *VisualFields         `gorm:"foreignKey:IDVisualFields;references:VisualFieldsID" json:"visual_fields"`
	TonometryEyes []TonometryEye        `gorm:"foreignKey:ExternalSleEyeID" json:"tonometry_eyes"`
}

func (ExternalSleEye) TableName() string { return "external_sle_eye" }
