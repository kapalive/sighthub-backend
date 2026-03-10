// internal/models/vision_exam/preliminary_eye_exam.go
package vision_exam

// PreliminaryEyeExam ↔ table: preliminary_eye_exam
type PreliminaryEyeExam struct {
	IDPreliminaryEyeExam       int64   `gorm:"column:id_preliminary_eye_exam;primaryKey;autoIncrement"         json:"id_preliminary_eye_exam"`
	EntranceGlassesID          *int64  `gorm:"column:entrance_glasses_id;uniqueIndex"                          json:"entrance_glasses_id,omitempty"`
	EntranceContLensID         *int64  `gorm:"column:entrance_cont_lens_id;uniqueIndex"                        json:"entrance_cont_lens_id,omitempty"`
	UnaidedVADistanceID        *int64  `gorm:"column:unaided_va_distance_id;uniqueIndex"                       json:"unaided_va_distance_id,omitempty"`
	UnaidedPHDistanceID        *int64  `gorm:"column:unaided_ph_distance_id;uniqueIndex"                       json:"unaided_ph_distance_id,omitempty"`
	UnaidedVANearID            *int64  `gorm:"column:unaided_va_near_id;uniqueIndex"                           json:"unaided_va_near_id,omitempty"`
	AidedVADistanceID          *int64  `gorm:"column:aided_va_distance_id;uniqueIndex"                         json:"aided_va_distance_id,omitempty"`
	AidedPHDistanceID          *int64  `gorm:"column:aided_ph_distance_id;uniqueIndex"                         json:"aided_ph_distance_id,omitempty"`
	AidedVANearID              *int64  `gorm:"column:aided_va_near_id;uniqueIndex"                             json:"aided_va_near_id,omitempty"`
	AidedByGlasses             bool    `gorm:"column:aided_by_glasses;not null;default:false"                  json:"aided_by_glasses"`
	AidedByContacts            bool    `gorm:"column:aided_by_contacts;not null;default:false"                 json:"aided_by_contacts"`
	ConfrontationID            *int64  `gorm:"column:confrontation_id;uniqueIndex"                             json:"confrontation_id,omitempty"`
	AutomatedID                *int64  `gorm:"column:automated_id;uniqueIndex"                                 json:"automated_id,omitempty"`
	MotilityID                 *int64  `gorm:"column:motility_id;uniqueIndex"                                  json:"motility_id,omitempty"`
	PupilsID                   *int64  `gorm:"column:pupils_id;uniqueIndex"                                    json:"pupils_id,omitempty"`
	ColorVisionID              *int64  `gorm:"column:color_vision_id;uniqueIndex"                              json:"color_vision_id,omitempty"`
	DistanceCoverTest          *string `gorm:"column:distance_cover_test;type:text"                            json:"distance_cover_test,omitempty"`
	NearCoverTest              *string `gorm:"column:near_cover_test;type:text"                                json:"near_cover_test,omitempty"`
	NpcTest                    *string `gorm:"column:npc_test;type:text"                                       json:"npc_test,omitempty"`
	BrucknerID                 *int64  `gorm:"column:bruckner_id;uniqueIndex"                                  json:"bruckner_id,omitempty"`
	AmslerGridID               *int64  `gorm:"column:amsler_grid_id;uniqueIndex"                               json:"amsler_grid_id,omitempty"`
	Worth4Dot                  *string `gorm:"column:worth_4_dot;type:text"                                    json:"worth_4_dot,omitempty"`
	StereoVision               *string `gorm:"column:stereo_vision;type:text"                                  json:"stereo_vision,omitempty"`
	FixationDisparity          *string `gorm:"column:fixation_disparity;type:text"                             json:"fixation_disparity,omitempty"`
	DistanceVonGraefePhorialID *int64  `gorm:"column:distance_von_graefe_phoria_id;uniqueIndex"                json:"distance_von_graefe_phoria_id,omitempty"`
	NearVonGraefePhorialID     *int64  `gorm:"column:near_von_graefe_phoria_id;uniqueIndex"                    json:"near_von_graefe_phoria_id,omitempty"`
	NearPointTestingID         *int64  `gorm:"column:near_point_testing_id"                                    json:"near_point_testing_id,omitempty"`
	AutorefractorPreliminaryID *int64  `gorm:"column:autorefractor_preliminary_id"                             json:"autorefractor_preliminary_id,omitempty"`
	AutoKeratometerPreliminaryID *int64 `gorm:"column:auto_keratometer_preliminary_id"                          json:"auto_keratometer_preliminary_id,omitempty"`
	BloodPressureID            *int    `gorm:"column:blood_pressure_id"                                        json:"blood_pressure_id,omitempty"`
	IrisColor                  string  `gorm:"column:iris_color;type:varchar(20);not null;default:'n/a'"       json:"iris_color"` // Blue|Green|Hazel|Brown|Heterochromia|n/a
	Note                       *string `gorm:"column:note;type:text"                                           json:"note,omitempty"`
	EyeExamID                  int64   `gorm:"column:eye_exam_id;not null"                                     json:"eye_exam_id"`

	// preload
	EntranceGlasses          *EntranceGlasses          `gorm:"foreignKey:EntranceGlassesID;references:IDEntranceGlasses"                   json:"-"`
	EntranceContLens         *EntranceContLens         `gorm:"foreignKey:EntranceContLensID;references:IDEntranceContLens"                 json:"-"`
	UnaidedVADistance        *UnaidedVADistance        `gorm:"foreignKey:UnaidedVADistanceID;references:IDUnaidedVADistance"               json:"-"`
	UnaidedPHDistance        *UnaidedPHDistance        `gorm:"foreignKey:UnaidedPHDistanceID;references:IDUnaidedPHDistance"               json:"-"`
	UnaidedVANear            *UnaidedVANear            `gorm:"foreignKey:UnaidedVANearID;references:IDUnaidedVANear"                       json:"-"`
	AidedVADistance          *AidedVADistance          `gorm:"foreignKey:AidedVADistanceID;references:IDAidedVADistance"                   json:"-"`
	AidedPHDistance          *AidedPHDistance          `gorm:"foreignKey:AidedPHDistanceID;references:IDAidedPHDistance"                   json:"-"`
	AidedVANear              *AidedVANear              `gorm:"foreignKey:AidedVANearID;references:IDAidedVANear"                           json:"-"`
	Confrontation            *Confrontation            `gorm:"foreignKey:ConfrontationID;references:IDConfrontation"                       json:"-"`
	Automated                *Automated                `gorm:"foreignKey:AutomatedID;references:IDAutomated"                               json:"-"`
	Motility                 *Motility                 `gorm:"foreignKey:MotilityID;references:IDMotility"                                 json:"-"`
	Pupils                   *Pupils                   `gorm:"foreignKey:PupilsID;references:IDPupils"                                     json:"-"`
	ColorVision              *ColorVision              `gorm:"foreignKey:ColorVisionID;references:IDColorVision"                           json:"-"`
	Bruckner                 *Bruckner                 `gorm:"foreignKey:BrucknerID;references:IDBruckner"                                 json:"-"`
	AmslerGrid               *AmslerGrid               `gorm:"foreignKey:AmslerGridID;references:IDAmslerGrid"                             json:"-"`
	DistanceVonGraefePhoria  *DistanceVonGraefePhoria  `gorm:"foreignKey:DistanceVonGraefePhorialID;references:IDDistanceVonGraefePhoria"  json:"-"`
	NearVonGraefePhoria      *NearVonGraefePhoria      `gorm:"foreignKey:NearVonGraefePhorialID;references:IDNearVonGraefePhoria"          json:"-"`
	NearPointTesting         *NearPointTesting         `gorm:"foreignKey:NearPointTestingID;references:IDNearPointTesting"                 json:"-"`
	AutorefractorPreliminary *AutorefractorPreliminary `gorm:"foreignKey:AutorefractorPreliminaryID;references:IDAutorefractorPreliminary" json:"-"`
	AutoKeratometerPreliminary *AutoKeratometerPreliminary `gorm:"foreignKey:AutoKeratometerPreliminaryID;references:IDAutoKeratometerPreliminary" json:"-"`
	BloodPressure            *BloodPressure            `gorm:"foreignKey:BloodPressureID;references:IDBloodPressure"                       json:"-"`
}

func (PreliminaryEyeExam) TableName() string { return "preliminary_eye_exam" }

func (p *PreliminaryEyeExam) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_preliminary_eye_exam":          p.IDPreliminaryEyeExam,
		"entrance_glasses_id":              p.EntranceGlassesID,
		"entrance_cont_lens_id":            p.EntranceContLensID,
		"unaided_va_distance_id":           p.UnaidedVADistanceID,
		"unaided_ph_distance_id":           p.UnaidedPHDistanceID,
		"unaided_va_near_id":               p.UnaidedVANearID,
		"aided_va_distance_id":             p.AidedVADistanceID,
		"aided_ph_distance_id":             p.AidedPHDistanceID,
		"aided_va_near_id":                 p.AidedVANearID,
		"aided_by_glasses":                 p.AidedByGlasses,
		"aided_by_contacts":                p.AidedByContacts,
		"confrontation_id":                 p.ConfrontationID,
		"automated_id":                     p.AutomatedID,
		"motility_id":                      p.MotilityID,
		"pupils_id":                        p.PupilsID,
		"color_vision_id":                  p.ColorVisionID,
		"distance_cover_test":              p.DistanceCoverTest,
		"near_cover_test":                  p.NearCoverTest,
		"npc_test":                         p.NpcTest,
		"bruckner_id":                      p.BrucknerID,
		"amsler_grid_id":                   p.AmslerGridID,
		"worth_4_dot":                      p.Worth4Dot,
		"stereo_vision":                    p.StereoVision,
		"fixation_disparity":               p.FixationDisparity,
		"distance_von_graefe_phoria_id":    p.DistanceVonGraefePhorialID,
		"near_von_graefe_phoria_id":        p.NearVonGraefePhorialID,
		"near_point_testing_id":            p.NearPointTestingID,
		"autorefractor_preliminary_id":     p.AutorefractorPreliminaryID,
		"auto_keratometer_preliminary_id":  p.AutoKeratometerPreliminaryID,
		"blood_pressure_id":               p.BloodPressureID,
		"iris_color":                       p.IrisColor,
		"note":                             p.Note,
		"eye_exam_id":                      p.EyeExamID,
	}
	return m
}
