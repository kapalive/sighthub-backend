// internal/models/vision_exam/slrp_and_referral.go
package vision_exam

import "time"

// SubjectiveSLRPEyeExam ↔ table: subjective_slrp_eye_exam
type SubjectiveSLRPEyeExam struct {
	IDSubjectiveSLRPEyeExam int64   `gorm:"column:id_subjective_slrp_eye_exam;primaryKey;autoIncrement" json:"id_subjective_slrp_eye_exam"`
	AccompaniedBy           *string `gorm:"column:accompanied_by;type:text"       json:"accompanied_by,omitempty"`
	SupervisingTechnician   *string `gorm:"column:supervising_technician;type:text" json:"supervising_technician,omitempty"`
	Changes                 *string `gorm:"column:changes;type:text"               json:"changes,omitempty"`
}
func (SubjectiveSLRPEyeExam) TableName() string { return "subjective_slrp_eye_exam" }
func (s *SubjectiveSLRPEyeExam) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_subjective_slrp_eye_exam": s.IDSubjectiveSLRPEyeExam,
		"accompanied_by": s.AccompaniedBy, "supervising_technician": s.SupervisingTechnician, "changes": s.Changes,
	}
}

// ObjectiveSLRPEyeExam ↔ table: objective_slrp_eye_exam
type ObjectiveSLRPEyeExam struct {
	IDObjectiveSLRPEyeExam         int64   `gorm:"column:id_objective_slrp_eye_exam;primaryKey;autoIncrement" json:"id_objective_slrp_eye_exam"`
	UtilHeadphonesStimulationAuSys bool    `gorm:"column:util_headphones_stimulation_au_sys;not null;default:false" json:"util_headphones_stimulation_au_sys"`
	ToleratedHeadphonesWell        bool    `gorm:"column:tolerated_headphones_well;not null;default:false"         json:"tolerated_headphones_well"`
	AttemptRemHeadphones           bool    `gorm:"column:attempt_rem_headphones;not null;default:false"            json:"attempt_rem_headphones"`
	CommentsObservations1          *string `gorm:"column:comments_observations1;type:text"                         json:"comments_observations1,omitempty"`
	UtilizMovTablBedStim           bool    `gorm:"column:utiliz_mov_tabl_bed_stim;not null;default:false"          json:"utiliz_mov_tabl_bed_stim"`
	TolerMoveWell                  bool    `gorm:"column:toler_move_well;not null;default:false"                   json:"toler_move_well"`
	AttempGetOffTab                bool    `gorm:"column:attemp_get_off_tab;not null;default:false"                json:"attemp_get_off_tab"`
	CommentsObservations2          *string `gorm:"column:comments_observations2;type:text"                         json:"comments_observations2,omitempty"`
	UtiliLightDeviceStimVs         bool    `gorm:"column:utili_light_device_stim_vs;not null;default:false"        json:"utili_light_device_stim_vs"`
	WavelengthsPresentedToday      *string `gorm:"column:wavelengths_presented_today;type:varchar(50)"             json:"wavelengths_presented_today,omitempty"`
	Magenta                        bool    `gorm:"column:magenta;not null;default:false" json:"magenta"`
	Ruby                           bool    `gorm:"column:ruby;not null;default:false"    json:"ruby"`
	Red                            bool    `gorm:"column:red;not null;default:false"     json:"red"`
	YellowGreen                    bool    `gorm:"column:yellow_green;not null;default:false" json:"yellow_green"`
	BlueGreen                      bool    `gorm:"column:blue_green;not null;default:false"   json:"blue_green"`
	Violet                         bool    `gorm:"column:violet;not null;default:false"  json:"violet"`
	TolerLightWell                 bool    `gorm:"column:toler_light_well;not null;default:false" json:"toler_light_well"`
	ClosedEyes                     bool    `gorm:"column:closed_eyes;not null;default:false"      json:"closed_eyes"`
	AttemptPushAwayLight           bool    `gorm:"column:attempt_push_away_light;not null;default:false" json:"attempt_push_away_light"`
	CommentsObservations3          *string `gorm:"column:comments_observations3;type:text" json:"comments_observations3,omitempty"`
}
func (ObjectiveSLRPEyeExam) TableName() string { return "objective_slrp_eye_exam" }
func (o *ObjectiveSLRPEyeExam) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_objective_slrp_eye_exam": o.IDObjectiveSLRPEyeExam}
}

// AssessmentSLRPEyeExam ↔ table: assessment_slrp_eye_exam
type AssessmentSLRPEyeExam struct {
	IDAssessmentSLRPEyeExam  int64   `gorm:"column:id_assessment_slrp_eye_exam;primaryKey;autoIncrement" json:"id_assessment_slrp_eye_exam"`
	ToleratedSessionWellToday bool   `gorm:"column:tolerated_session_well_today;not null;default:false"   json:"tolerated_session_well_today"`
	Comments                  *string `gorm:"column:comments;type:text"                                    json:"comments,omitempty"`
}
func (AssessmentSLRPEyeExam) TableName() string { return "assessment_slrp_eye_exam" }
func (a *AssessmentSLRPEyeExam) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_assessment_slrp_eye_exam": a.IDAssessmentSLRPEyeExam, "tolerated_session_well_today": a.ToleratedSessionWellToday, "comments": a.Comments}
}

// PlanSLRPEyeExam ↔ table: plan_slrp_eye_exam
type PlanSLRPEyeExam struct {
	IDPlanSLRPEyeExam       int64   `gorm:"column:id_plan_slrp_eye_exam;primaryKey;autoIncrement" json:"id_plan_slrp_eye_exam"`
	ContinueProgramScheduled bool   `gorm:"column:continue_program_scheduled;not null;default:false" json:"continue_program_scheduled"`
	ModifyProgram            *string `gorm:"column:modify_program;type:text"                          json:"modify_program,omitempty"`
}
func (PlanSLRPEyeExam) TableName() string { return "plan_slrp_eye_exam" }
func (p *PlanSLRPEyeExam) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_plan_slrp_eye_exam": p.IDPlanSLRPEyeExam, "continue_program_scheduled": p.ContinueProgramScheduled, "modify_program": p.ModifyProgram}
}

// SlrpEyeExam ↔ table: slrp_eye_exam
type SlrpEyeExam struct {
	IDSlrpEyeExam             int64      `gorm:"column:id_slrp_eye_exam;primaryKey;autoIncrement" json:"id_slrp_eye_exam"`
	EyeExamID                 int64      `gorm:"column:eye_exam_id;not null"                       json:"eye_exam_id"`
	SubjectiveSLRPEyeExamID   *int64     `gorm:"column:subjective_slrp_eye_exam_id"                json:"subjective_slrp_eye_exam_id,omitempty"`
	ObjectiveSLRPEyeExamID    *int64     `gorm:"column:objective_slrp_eye_exam_id"                 json:"objective_slrp_eye_exam_id,omitempty"`
	AssessmentSLRPEyeExamID   *int64     `gorm:"column:assessment_slrp_eye_exam_id"                json:"assessment_slrp_eye_exam_id,omitempty"`
	PlanSLRPEyeExamID         *int64     `gorm:"column:plan_slrp_eye_exam_id"                      json:"plan_slrp_eye_exam_id,omitempty"`
	StartDate                 *time.Time `gorm:"column:start_date;type:date"                       json:"start_date,omitempty"`
	StartTime                 *string    `gorm:"column:start_time;type:time"                       json:"start_time,omitempty"`
	EndDate                   *time.Time `gorm:"column:end_date;type:date"                         json:"end_date,omitempty"`
	EndTime                   *string    `gorm:"column:end_time;type:time"                         json:"end_time,omitempty"`

	Subjective  *SubjectiveSLRPEyeExam  `gorm:"foreignKey:SubjectiveSLRPEyeExamID;references:IDSubjectiveSLRPEyeExam"   json:"-"`
	Objective   *ObjectiveSLRPEyeExam   `gorm:"foreignKey:ObjectiveSLRPEyeExamID;references:IDObjectiveSLRPEyeExam"     json:"-"`
	Assessment  *AssessmentSLRPEyeExam  `gorm:"foreignKey:AssessmentSLRPEyeExamID;references:IDAssessmentSLRPEyeExam"   json:"-"`
	Plan        *PlanSLRPEyeExam        `gorm:"foreignKey:PlanSLRPEyeExamID;references:IDPlanSLRPEyeExam"               json:"-"`
}
func (SlrpEyeExam) TableName() string { return "slrp_eye_exam" }
func (s *SlrpEyeExam) ToMap() map[string]interface{} {
	m := map[string]interface{}{"id_slrp_eye_exam": s.IDSlrpEyeExam, "eye_exam_id": s.EyeExamID}
	if s.StartDate != nil { m["start_date"] = s.StartDate.Format("2006-01-02") } else { m["start_date"] = nil }
	if s.EndDate != nil { m["end_date"] = s.EndDate.Format("2006-01-02") } else { m["end_date"] = nil }
	return m
}

// ReferralDoctor ↔ table: referral_doctor
type ReferralDoctor struct {
	IDReferralDoctor int    `gorm:"column:id_referral_doctor;primaryKey;autoIncrement" json:"id_referral_doctor"`
	Salutation       *string `gorm:"column:salutation;type:varchar(4)"    json:"salutation,omitempty"`
	NPI              *string `gorm:"column:npi;type:varchar(20)"          json:"npi,omitempty"`
	LastName         *string `gorm:"column:last_name;type:text"           json:"last_name,omitempty"`
	FirstName        *string `gorm:"column:first_name;type:text"          json:"first_name,omitempty"`
	Address          *string `gorm:"column:address;type:text"             json:"address,omitempty"`
	Address2         *string `gorm:"column:address2;type:varchar(20)"     json:"address2,omitempty"`
	City             *string `gorm:"column:city;type:varchar(100)"        json:"city,omitempty"`
	State            *string `gorm:"column:state;type:varchar(2)"         json:"state,omitempty"`
	Zip              *string `gorm:"column:zip;type:varchar(6)"           json:"zip,omitempty"`
	Phone            *string `gorm:"column:phone;type:varchar(16)"        json:"phone,omitempty"`
	Fax              *string `gorm:"column:fax;type:varchar(16)"          json:"fax,omitempty"`
	Email            *string `gorm:"column:email;type:varchar(50)"        json:"email,omitempty"`
}
func (ReferralDoctor) TableName() string { return "referral_doctor" }
func (r *ReferralDoctor) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_referral_doctor": r.IDReferralDoctor, "salutation": r.Salutation,
		"npi": r.NPI, "last_name": r.LastName, "first_name": r.FirstName,
		"phone": r.Phone, "fax": r.Fax, "email": r.Email,
	}
}

// ReferralLetter ↔ table: referral_letter
type ReferralLetter struct {
	IDReferralLetter      int64  `gorm:"column:id_referral_letter;primaryKey;autoIncrement" json:"id_referral_letter"`
	TitleLetter           *string `gorm:"column:title_letter;type:text"       json:"title_letter,omitempty"`
	IntroLetter           *string `gorm:"column:intro_letter;type:text"       json:"intro_letter,omitempty"`
	TestsLetter           *string `gorm:"column:tests_letter;type:text"       json:"tests_letter,omitempty"`
	IssueLetter           *string `gorm:"column:issue_letter;type:text"       json:"issue_letter,omitempty"`
	EyeExamID             int64  `gorm:"column:eye_exam_id;not null"         json:"eye_exam_id"`
	ToReferralDoctorID    *int64 `gorm:"column:to_referral_doctor_id"        json:"to_referral_doctor_id,omitempty"`
	CcReferralDoctorID    *int64 `gorm:"column:cc_referral_doctor_id"        json:"cc_referral_doctor_id,omitempty"`

	ToDoctor *ReferralDoctor `gorm:"foreignKey:ToReferralDoctorID;references:IDReferralDoctor" json:"-"`
	CcDoctor *ReferralDoctor `gorm:"foreignKey:CcReferralDoctorID;references:IDReferralDoctor" json:"-"`
}
func (ReferralLetter) TableName() string { return "referral_letter" }
func (r *ReferralLetter) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_referral_letter": r.IDReferralLetter, "eye_exam_id": r.EyeExamID,
		"title_letter": r.TitleLetter, "intro_letter": r.IntroLetter,
		"tests_letter": r.TestsLetter, "issue_letter": r.IssueLetter,
		"to_referral_doctor_id": r.ToReferralDoctorID, "cc_referral_doctor_id": r.CcReferralDoctorID,
	}
}
