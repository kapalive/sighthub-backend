package slrp

type ObjectiveSLRPEyeExam struct {
	IDObjectiveSLRPEyeExam       int64   `gorm:"column:id_objective_slrp_eye_exam;primaryKey;autoIncrement" json:"id_objective_slrp_eye_exam"`
	UtilHeadphonesStimulationAuSys *bool `gorm:"column:util_headphones_stimulation_au_sys;default:false" json:"util_headphones_stimulation_au_sys"`
	ToleratedHeadphonesWell       *bool  `gorm:"column:tolerated_headphones_well;default:false" json:"tolerated_headphones_well"`
	AttemptRemHeadphones          *bool  `gorm:"column:attempt_rem_headphones;default:false" json:"attempt_rem_headphones"`
	CommentsObservations1         *string `gorm:"column:comments_observations1;type:text" json:"comments_observations1"`
	UtilizMovTablBedStim          *bool  `gorm:"column:utiliz_mov_tabl_bed_stim;default:false" json:"utiliz_mov_tabl_bed_stim"`
	TolerMoveWell                 *bool  `gorm:"column:toler_move_well;default:false" json:"toler_move_well"`
	AttempGetOffTab               *bool  `gorm:"column:attemp_get_off_tab;default:false" json:"attemp_get_off_tab"`
	CommentsObservations2         *string `gorm:"column:comments_observations2;type:text" json:"comments_observations2"`
	UtiliLightDeviceStimVs        *bool  `gorm:"column:utili_light_device_stim_vs;default:false" json:"utili_light_device_stim_vs"`
	WavelengthsPresentedToday     *string `gorm:"column:wavelengths_presented_today;size:50" json:"wavelengths_presented_today"`
	Magenta                       *bool  `gorm:"column:magenta;default:false" json:"magenta"`
	Ruby                          *bool  `gorm:"column:ruby;default:false" json:"ruby"`
	Red                           *bool  `gorm:"column:red;default:false" json:"red"`
	YellowGreen                   *bool  `gorm:"column:yellow_green;default:false" json:"yellow_green"`
	BlueGreen                     *bool  `gorm:"column:blue_green;default:false" json:"blue_green"`
	Violet                        *bool  `gorm:"column:violet;default:false" json:"violet"`
	TolerLightWell                *bool  `gorm:"column:toler_light_well;default:false" json:"toler_light_well"`
	ClosedEyes                    *bool  `gorm:"column:closed_eyes;default:false" json:"closed_eyes"`
	AttemptPushAwayLight          *bool  `gorm:"column:attempt_push_away_light;default:false" json:"attempt_push_away_light"`
	CommentsObservations3         *string `gorm:"column:comments_observations3;type:text" json:"comments_observations3"`
}

func (ObjectiveSLRPEyeExam) TableName() string { return "objective_slrp_eye_exam" }
