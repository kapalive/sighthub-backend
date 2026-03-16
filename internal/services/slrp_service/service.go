package slrp_service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	authModel  "sighthub-backend/internal/models/auth"
	empModel   "sighthub-backend/internal/models/employees"
	slrpModel  "sighthub-backend/internal/models/medical/vision_exam/slrp"
	visionModel "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── helpers ──────────────────────────────────────────────────────────────────

func (s *Service) getEmployee(username string) (*empModel.Employee, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	return &emp, nil
}

func parseDate(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

// ─── input types ──────────────────────────────────────────────────────────────

type SubjectiveInput struct {
	AccompaniedBy         *string `json:"accompanied_by"`
	SupervisingTechnician *string `json:"supervising_technician"`
	Changes               *string `json:"changes"`
}

type ObjectiveInput struct {
	UtilHeadphonesStimulationAuSys *bool   `json:"util_headphones_stimulation_au_sys"`
	ToleratedHeadphonesWell        *bool   `json:"tolerated_headphones_well"`
	AttemptRemHeadphones           *bool   `json:"attempt_rem_headphones"`
	CommentsObservations1          *string `json:"comments_observations1"`
	UtilizMovTablBedStim           *bool   `json:"utiliz_mov_tabl_bed_stim"`
	TolerMoveWell                  *bool   `json:"toler_move_well"`
	AttempGetOffTab                *bool   `json:"attemp_get_off_tab"`
	CommentsObservations2          *string `json:"comments_observations2"`
	UtiliLightDeviceStimVs         *bool   `json:"utili_light_device_stim_vs"`
	WavelengthsPresentedToday      *string `json:"wavelengths_presented_today"`
	Magenta                        *bool   `json:"magenta"`
	Ruby                           *bool   `json:"ruby"`
	Red                            *bool   `json:"red"`
	YellowGreen                    *bool   `json:"yellow_green"`
	BlueGreen                      *bool   `json:"blue_green"`
	Violet                         *bool   `json:"violet"`
	TolerLightWell                 *bool   `json:"toler_light_well"`
	ClosedEyes                     *bool   `json:"closed_eyes"`
	AttemptPushAwayLight           *bool   `json:"attempt_push_away_light"`
	CommentsObservations3          *string `json:"comments_observations3"`
}

type AssessmentInput struct {
	ToleratedSessionWellToday *bool   `json:"tolerated_session_well_today"`
	Comments                  *string `json:"comments"`
}

type PlanInput struct {
	ContinueProgramScheduled *bool   `json:"continue_program_scheduled"`
	ModifyProgram            *string `json:"modify_program"`
}

type CreateSLRPInput struct {
	StartDate  *string          `json:"start_date"`
	StartTime  *string          `json:"start_time"`
	EndDate    *string          `json:"end_date"`
	EndTime    *string          `json:"end_time"`
	Subjective *SubjectiveInput `json:"subjective"`
	Objective  *ObjectiveInput  `json:"objective"`
	Assessment *AssessmentInput `json:"assessment"`
	Plan       *PlanInput       `json:"plan"`
}

// ─── result helpers ───────────────────────────────────────────────────────────

func subjectiveToMap(r *slrpModel.SubjectiveSLRPEyeExam) map[string]interface{} {
	if r == nil {
		return nil
	}
	return map[string]interface{}{
		"id_subjective_slrp_eye_exam": r.IDSubjectiveSLRPEyeExam,
		"accompanied_by":              r.AccompaniedBy,
		"supervising_technician":      r.SupervisingTechnician,
		"changes":                     r.Changes,
	}
}

func objectiveToMap(r *slrpModel.ObjectiveSLRPEyeExam) map[string]interface{} {
	if r == nil {
		return nil
	}
	return map[string]interface{}{
		"id_objective_slrp_eye_exam":       r.IDObjectiveSLRPEyeExam,
		"util_headphones_stimulation_au_sys": r.UtilHeadphonesStimulationAuSys,
		"tolerated_headphones_well":         r.ToleratedHeadphonesWell,
		"attempt_rem_headphones":            r.AttemptRemHeadphones,
		"comments_observations1":            r.CommentsObservations1,
		"utiliz_mov_tabl_bed_stim":          r.UtilizMovTablBedStim,
		"toler_move_well":                   r.TolerMoveWell,
		"attemp_get_off_tab":                r.AttempGetOffTab,
		"comments_observations2":            r.CommentsObservations2,
		"utili_light_device_stim_vs":        r.UtiliLightDeviceStimVs,
		"wavelengths_presented_today":       r.WavelengthsPresentedToday,
		"magenta":                           r.Magenta,
		"ruby":                              r.Ruby,
		"red":                               r.Red,
		"yellow_green":                      r.YellowGreen,
		"blue_green":                        r.BlueGreen,
		"violet":                            r.Violet,
		"toler_light_well":                  r.TolerLightWell,
		"closed_eyes":                       r.ClosedEyes,
		"attempt_push_away_light":           r.AttemptPushAwayLight,
		"comments_observations3":            r.CommentsObservations3,
	}
}

func assessmentToMap(r *slrpModel.AssessmentSLRPEyeExam) map[string]interface{} {
	if r == nil {
		return nil
	}
	return map[string]interface{}{
		"id_assessment_slrp_eye_exam":  r.IDAssessmentSLRPEyeExam,
		"tolerated_session_well_today": r.ToleratedSessionWellToday,
		"comments":                     r.Comments,
	}
}

func planToMap(r *slrpModel.PlanSLRPEyeExam) map[string]interface{} {
	if r == nil {
		return nil
	}
	return map[string]interface{}{
		"id_plan_slrp_eye_exam":      r.IDPlanSLRPEyeExam,
		"continue_program_scheduled": r.ContinueProgramScheduled,
		"modify_program":             r.ModifyProgram,
	}
}

func (s *Service) buildResult(examID int64, slrp *slrpModel.SLRPEyeExam) map[string]interface{} {
	var subj *slrpModel.SubjectiveSLRPEyeExam
	var obj  *slrpModel.ObjectiveSLRPEyeExam
	var ass  *slrpModel.AssessmentSLRPEyeExam
	var plan *slrpModel.PlanSLRPEyeExam

	if slrp.SubjectiveSLRPEyeExamID != nil {
		var r slrpModel.SubjectiveSLRPEyeExam
		if s.db.First(&r, *slrp.SubjectiveSLRPEyeExamID).Error == nil {
			subj = &r
		}
	}
	if slrp.ObjectiveSLRPEyeExamID != nil {
		var r slrpModel.ObjectiveSLRPEyeExam
		if s.db.First(&r, *slrp.ObjectiveSLRPEyeExamID).Error == nil {
			obj = &r
		}
	}
	if slrp.AssessmentSLRPEyeExamID != nil {
		var r slrpModel.AssessmentSLRPEyeExam
		if s.db.First(&r, *slrp.AssessmentSLRPEyeExamID).Error == nil {
			ass = &r
		}
	}
	if slrp.PlanSLRPEyeExamID != nil {
		var r slrpModel.PlanSLRPEyeExam
		if s.db.First(&r, *slrp.PlanSLRPEyeExamID).Error == nil {
			plan = &r
		}
	}

	return map[string]interface{}{
		"id_slrp_eye_exam": slrp.IDSLRPEyeExam,
		"start_date":       formatDate(slrp.StartDate),
		"start_time":       slrp.StartTime,
		"end_date":         formatDate(slrp.EndDate),
		"end_time":         slrp.EndTime,
		"subjective":       subjectiveToMap(subj),
		"objective":        objectiveToMap(obj),
		"assessment":       assessmentToMap(ass),
		"plan":             planToMap(plan),
	}
}

// ─── service methods ──────────────────────────────────────────────────────────

func (s *Service) CreateSLRP(examID int64, input CreateSLRPInput) (map[string]interface{}, error) {
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	slrp := slrpModel.SLRPEyeExam{
		EyeExamID: examID,
		StartDate: parseDate(input.StartDate),
		StartTime: input.StartTime,
		EndDate:   parseDate(input.EndDate),
		EndTime:   input.EndTime,
	}

	if input.Subjective != nil {
		subj := slrpModel.SubjectiveSLRPEyeExam{
			AccompaniedBy:         input.Subjective.AccompaniedBy,
			SupervisingTechnician: input.Subjective.SupervisingTechnician,
			Changes:               input.Subjective.Changes,
		}
		if err := tx.Create(&subj).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		slrp.SubjectiveSLRPEyeExamID = &subj.IDSubjectiveSLRPEyeExam
	}

	if input.Objective != nil {
		obj := slrpModel.ObjectiveSLRPEyeExam{
			UtilHeadphonesStimulationAuSys: input.Objective.UtilHeadphonesStimulationAuSys,
			ToleratedHeadphonesWell:        input.Objective.ToleratedHeadphonesWell,
			AttemptRemHeadphones:           input.Objective.AttemptRemHeadphones,
			CommentsObservations1:          input.Objective.CommentsObservations1,
			UtilizMovTablBedStim:           input.Objective.UtilizMovTablBedStim,
			TolerMoveWell:                  input.Objective.TolerMoveWell,
			AttempGetOffTab:                input.Objective.AttempGetOffTab,
			CommentsObservations2:          input.Objective.CommentsObservations2,
			UtiliLightDeviceStimVs:         input.Objective.UtiliLightDeviceStimVs,
			WavelengthsPresentedToday:      input.Objective.WavelengthsPresentedToday,
			Magenta:                        input.Objective.Magenta,
			Ruby:                           input.Objective.Ruby,
			Red:                            input.Objective.Red,
			YellowGreen:                    input.Objective.YellowGreen,
			BlueGreen:                      input.Objective.BlueGreen,
			Violet:                         input.Objective.Violet,
			TolerLightWell:                 input.Objective.TolerLightWell,
			ClosedEyes:                     input.Objective.ClosedEyes,
			AttemptPushAwayLight:           input.Objective.AttemptPushAwayLight,
			CommentsObservations3:          input.Objective.CommentsObservations3,
		}
		if err := tx.Create(&obj).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		slrp.ObjectiveSLRPEyeExamID = &obj.IDObjectiveSLRPEyeExam
	}

	if input.Assessment != nil {
		ass := slrpModel.AssessmentSLRPEyeExam{
			ToleratedSessionWellToday: input.Assessment.ToleratedSessionWellToday,
			Comments:                  input.Assessment.Comments,
		}
		if err := tx.Create(&ass).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		slrp.AssessmentSLRPEyeExamID = &ass.IDAssessmentSLRPEyeExam
	}

	if input.Plan != nil {
		pl := slrpModel.PlanSLRPEyeExam{
			ContinueProgramScheduled: input.Plan.ContinueProgramScheduled,
			ModifyProgram:            input.Plan.ModifyProgram,
		}
		if err := tx.Create(&pl).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		slrp.PlanSLRPEyeExamID = &pl.IDPlanSLRPEyeExam
	}

	if err := tx.Create(&slrp).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	activitylog.Log(tx, "exam_slrp", "save", activitylog.WithEntity(examID))
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.buildResult(examID, &slrp), nil
}

func (s *Service) UpdateSLRP(username string, examID int64, input CreateSLRPInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.Passed {
		return nil, errors.New("exam has already been completed")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized to update this exam")
	}

	var slrp slrpModel.SLRPEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&slrp).Error; err != nil {
		return nil, errors.New("SLRP eye exam not found")
	}

	updates := map[string]interface{}{}
	if input.StartDate != nil {
		updates["start_date"] = parseDate(input.StartDate)
	}
	if input.StartTime != nil {
		updates["start_time"] = input.StartTime
	}
	if input.EndDate != nil {
		updates["end_date"] = parseDate(input.EndDate)
	}
	if input.EndTime != nil {
		updates["end_time"] = input.EndTime
	}
	if len(updates) > 0 {
		s.db.Model(&slrp).Updates(updates)
	}

	if input.Subjective != nil && slrp.SubjectiveSLRPEyeExamID != nil {
		var subj slrpModel.SubjectiveSLRPEyeExam
		if s.db.First(&subj, *slrp.SubjectiveSLRPEyeExamID).Error == nil {
			subjUpdates := map[string]interface{}{}
			if input.Subjective.AccompaniedBy != nil {
				subjUpdates["accompanied_by"] = input.Subjective.AccompaniedBy
			}
			if input.Subjective.SupervisingTechnician != nil {
				subjUpdates["supervising_technician"] = input.Subjective.SupervisingTechnician
			}
			if input.Subjective.Changes != nil {
				subjUpdates["changes"] = input.Subjective.Changes
			}
			if len(subjUpdates) > 0 {
				s.db.Model(&subj).Updates(subjUpdates)
			}
		}
	}

	if input.Objective != nil && slrp.ObjectiveSLRPEyeExamID != nil {
		var obj slrpModel.ObjectiveSLRPEyeExam
		if s.db.First(&obj, *slrp.ObjectiveSLRPEyeExamID).Error == nil {
			objUpdates := map[string]interface{}{}
			if input.Objective.UtilHeadphonesStimulationAuSys != nil {
				objUpdates["util_headphones_stimulation_au_sys"] = input.Objective.UtilHeadphonesStimulationAuSys
			}
			if input.Objective.ToleratedHeadphonesWell != nil {
				objUpdates["tolerated_headphones_well"] = input.Objective.ToleratedHeadphonesWell
			}
			if input.Objective.AttemptRemHeadphones != nil {
				objUpdates["attempt_rem_headphones"] = input.Objective.AttemptRemHeadphones
			}
			if input.Objective.CommentsObservations1 != nil {
				objUpdates["comments_observations1"] = input.Objective.CommentsObservations1
			}
			if input.Objective.UtilizMovTablBedStim != nil {
				objUpdates["utiliz_mov_tabl_bed_stim"] = input.Objective.UtilizMovTablBedStim
			}
			if input.Objective.TolerMoveWell != nil {
				objUpdates["toler_move_well"] = input.Objective.TolerMoveWell
			}
			if input.Objective.AttempGetOffTab != nil {
				objUpdates["attemp_get_off_tab"] = input.Objective.AttempGetOffTab
			}
			if input.Objective.CommentsObservations2 != nil {
				objUpdates["comments_observations2"] = input.Objective.CommentsObservations2
			}
			if input.Objective.UtiliLightDeviceStimVs != nil {
				objUpdates["utili_light_device_stim_vs"] = input.Objective.UtiliLightDeviceStimVs
			}
			if input.Objective.WavelengthsPresentedToday != nil {
				objUpdates["wavelengths_presented_today"] = input.Objective.WavelengthsPresentedToday
			}
			if input.Objective.Magenta != nil {
				objUpdates["magenta"] = input.Objective.Magenta
			}
			if input.Objective.Ruby != nil {
				objUpdates["ruby"] = input.Objective.Ruby
			}
			if input.Objective.Red != nil {
				objUpdates["red"] = input.Objective.Red
			}
			if input.Objective.YellowGreen != nil {
				objUpdates["yellow_green"] = input.Objective.YellowGreen
			}
			if input.Objective.BlueGreen != nil {
				objUpdates["blue_green"] = input.Objective.BlueGreen
			}
			if input.Objective.Violet != nil {
				objUpdates["violet"] = input.Objective.Violet
			}
			if input.Objective.TolerLightWell != nil {
				objUpdates["toler_light_well"] = input.Objective.TolerLightWell
			}
			if input.Objective.ClosedEyes != nil {
				objUpdates["closed_eyes"] = input.Objective.ClosedEyes
			}
			if input.Objective.AttemptPushAwayLight != nil {
				objUpdates["attempt_push_away_light"] = input.Objective.AttemptPushAwayLight
			}
			if input.Objective.CommentsObservations3 != nil {
				objUpdates["comments_observations3"] = input.Objective.CommentsObservations3
			}
			if len(objUpdates) > 0 {
				s.db.Model(&obj).Updates(objUpdates)
			}
		}
	}

	if input.Assessment != nil && slrp.AssessmentSLRPEyeExamID != nil {
		var ass slrpModel.AssessmentSLRPEyeExam
		if s.db.First(&ass, *slrp.AssessmentSLRPEyeExamID).Error == nil {
			assUpdates := map[string]interface{}{}
			if input.Assessment.ToleratedSessionWellToday != nil {
				assUpdates["tolerated_session_well_today"] = input.Assessment.ToleratedSessionWellToday
			}
			if input.Assessment.Comments != nil {
				assUpdates["comments"] = input.Assessment.Comments
			}
			if len(assUpdates) > 0 {
				s.db.Model(&ass).Updates(assUpdates)
			}
		}
	}

	if input.Plan != nil && slrp.PlanSLRPEyeExamID != nil {
		var pl slrpModel.PlanSLRPEyeExam
		if s.db.First(&pl, *slrp.PlanSLRPEyeExamID).Error == nil {
			planUpdates := map[string]interface{}{}
			if input.Plan.ContinueProgramScheduled != nil {
				planUpdates["continue_program_scheduled"] = input.Plan.ContinueProgramScheduled
			}
			if input.Plan.ModifyProgram != nil {
				planUpdates["modify_program"] = input.Plan.ModifyProgram
			}
			if len(planUpdates) > 0 {
				s.db.Model(&pl).Updates(planUpdates)
			}
		}
	}

	activitylog.Log(s.db, "exam_slrp", "update", activitylog.WithEntity(examID))

	// reload updated slrp
	s.db.Where("eye_exam_id = ?", examID).First(&slrp)
	return s.buildResult(examID, &slrp), nil
}

func (s *Service) GetSLRP(examID int64) (map[string]interface{}, error) {
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}

	var slrp slrpModel.SLRPEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&slrp).Error; err != nil {
		return map[string]interface{}{
			"exam_id":    examID,
			"exists":     false,
			"start_date": nil,
			"start_time": nil,
			"end_date":   nil,
			"end_time":   nil,
			"subjective": nil,
			"objective":  nil,
			"assessment": nil,
			"plan":       nil,
		}, nil
	}

	result := s.buildResult(examID, &slrp)
	result["exam_id"] = examID
	result["exists"] = true
	return result, nil
}
