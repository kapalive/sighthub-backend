package appointment

import "fmt"

type ReasonsVisionProviderAppointment struct {
	IDReasonsVisionProviderAppointment int     `gorm:"column:id_reasons_vision_provider_appointment;primaryKey" json:"id_reasons_vision_provider_appointment"`
	Reason                             string  `gorm:"column:reason;type:text;not null"                                                         json:"reason"`
	Description                        *string `gorm:"column:description;type:text"                                                             json:"description,omitempty"`
}

func (ReasonsVisionProviderAppointment) TableName() string {
	return "reasons_vision_provider_appointment"
}

func (r *ReasonsVisionProviderAppointment) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_reasons_vision_provider_appointment": r.IDReasonsVisionProviderAppointment,
		"reason":                                 r.Reason,
		"description":                            r.Description,
	}
}

func (r *ReasonsVisionProviderAppointment) String() string {
	return fmt.Sprintf("<ReasonsVisionProviderAppointment %d: %s>", r.IDReasonsVisionProviderAppointment, r.Reason)
}
