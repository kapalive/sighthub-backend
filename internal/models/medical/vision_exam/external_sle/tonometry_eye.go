package external_sle

import "time"

type TonometryEye struct {
	IDTonometryEye   int64      `gorm:"column:id_tonometry_eye;primaryKey;autoIncrement" json:"id_tonometry_eye"`
	ExternalSleEyeID int64      `gorm:"column:external_sle_eye_id;not null" json:"external_sle_eye_id"`
	MethodTonometry  *string    `gorm:"column:method_tonometry;size:100" json:"method_tonometry"`
	DateTonometryEye *time.Time `gorm:"column:date_tonometry_eye;type:date" json:"date_tonometry_eye"`
	TimeTonometryEye *string    `gorm:"column:time_tonometry_eye;type:time" json:"time_tonometry_eye"`
	OdTonometryEye   *string    `gorm:"column:od_tonometry_eye;size:6" json:"od_tonometry_eye"`
	OsTonometryEye   *string    `gorm:"column:os_tonometry_eye;size:6" json:"os_tonometry_eye"`
}

func (TonometryEye) TableName() string { return "tonometry_eye" }
