package assessment_service

import (
	"errors"

	"gorm.io/gorm"

	empLoginModel "sighthub-backend/internal/models/auth"
	diseaseModel "sighthub-backend/internal/models/diseases"
	generalModel "sighthub-backend/internal/models/general"
	assessmentModel "sighthub-backend/internal/models/medical/vision_exam/assessment"
	empModel "sighthub-backend/internal/models/employees"
	visionModel "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ---------- input types ----------

type DiagnosisInput struct {
	Code    *string `json:"code"`
	LevelID *int64  `json:"level_id"`
	Type    *string `json:"type"`
	Title   *string `json:"title"`
}

type PQRSInput struct {
	IDPqrs int64 `json:"id_pqrs"`
}

type AssessmentInput struct {
	IDAssessmentEye *int64           `json:"id_assessment_eye"`
	Impression      *string          `json:"impression"`
	Plan            *string          `json:"plan"`
	Diagnosis       []DiagnosisInput `json:"diagnosis"`
	PQRS            []PQRSInput      `json:"pqrs"`
}

type SaveAssessmentInput struct {
	Assessments []AssessmentInput `json:"assessments"`
}

type UpdateAssessmentInput struct {
	Assessments []AssessmentInput `json:"assessments"`
}

type AddTopDiseaseInput struct {
	LevelID  int64   `json:"level_id"`
	Type     string  `json:"type"`
	Code     string  `json:"code"`
	Title    string  `json:"title"`
	GroupSet *string `json:"group_set"`
}

// ---------- result types ----------

type DiagnosisResult struct {
	IDAssessmentDiagnosis int64   `json:"id_assessment_diagnosis"`
	Code                  *string `json:"code"`
	LevelID               *int64  `json:"level_id"`
	Type                  *string `json:"type"`
	Title                 *string `json:"title"`
}

type PQRSResult struct {
	IDAssessmentPQRS int64   `json:"id_assessment_pqrs"`
	IDPqrs           int64   `json:"id_pqrs"`
	Code             *string `json:"code"`
	Title            *string `json:"title"`
}

type AssessmentResult struct {
	IDAssessmentEye int64             `json:"id_assessment_eye"`
	Impression      *string           `json:"impression"`
	Plan            *string           `json:"plan"`
	Diagnosis       []DiagnosisResult `json:"diagnosis"`
	PQRS            []PQRSResult      `json:"pqrs"`
}

type SearchResult struct {
	LevelID int64  `json:"level_id"`
	Code    string `json:"code"`
	Title   string `json:"title"`
	Type    string `json:"type"`
}

type DiseaseChildResult struct {
	LevelID int64  `json:"level_id"`
	Code    string `json:"code"`
	Title   string `json:"title"`
	Type    string `json:"type"`
}

type TopDiseaseResult struct {
	LevelID  int64                `json:"level_id"`
	Code     string               `json:"code"`
	Title    string               `json:"title"`
	Type     string               `json:"type"`
	GroupSet *string              `json:"group_set"`
	Children []DiseaseChildResult `json:"children"`
}

type PQRSListResult struct {
	IDPqrs int64   `json:"id_pqrs"`
	Code   string  `json:"code"`
	Title  *string `json:"title"`
	Group  *string `json:"group"`
}

// ---------- helpers ----------

func (s *Service) getEmployee(username string) (*empModel.Employee, error) {
	var login empLoginModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	return &emp, nil
}

func (s *Service) validateExam(emp *empModel.Employee, examID int64) error {
	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return errors.New("not authorized")
	}
	if exam.Passed {
		return errors.New("cannot update: exam is completed")
	}
	return nil
}

func (s *Service) buildAssessmentResult(db *gorm.DB, ae assessmentModel.AssessmentEye) AssessmentResult {
	var diagnoses []assessmentModel.AssessmentDiagnosis
	db.Where("assessment_eye_id = ?", ae.IDAssessmentEye).Find(&diagnoses)

	var pqrsLinks []assessmentModel.AssessmentPQRS
	db.Where("assessment_eye_id = ?", ae.IDAssessmentEye).Find(&pqrsLinks)

	diagResults := make([]DiagnosisResult, 0, len(diagnoses))
	for _, d := range diagnoses {
		diagResults = append(diagResults, DiagnosisResult{
			IDAssessmentDiagnosis: d.IDAssessmentDiagnosis,
			Code:                  d.Code,
			LevelID:               d.LevelID,
			Type:                  d.Type,
			Title:                 d.Title,
		})
	}

	pqrsResults := make([]PQRSResult, 0, len(pqrsLinks))
	for _, p := range pqrsLinks {
		var pqrs generalModel.PQRS
		var code *string
		var title *string
		if err := db.First(&pqrs, p.PqrsID).Error; err == nil {
			c := pqrs.Code
			code = &c
			title = pqrs.Title
		}
		pqrsResults = append(pqrsResults, PQRSResult{
			IDAssessmentPQRS: p.IDAssessmentPQRS,
			IDPqrs:           p.PqrsID,
			Code:             code,
			Title:            title,
		})
	}

	return AssessmentResult{
		IDAssessmentEye: ae.IDAssessmentEye,
		Impression:      ae.Impression,
		Plan:            ae.Plan,
		Diagnosis:       diagResults,
		PQRS:            pqrsResults,
	}
}

func (s *Service) insertDiagnoses(tx *gorm.DB, assessmentEyeID int64, items []DiagnosisInput) {
	for _, d := range items {
		if d.Code == nil || d.LevelID == nil || d.Type == nil {
			continue
		}
		tx.Create(&assessmentModel.AssessmentDiagnosis{
			AssessmentEyeID: assessmentEyeID,
			Code:            d.Code,
			LevelID:         d.LevelID,
			Type:            d.Type,
			Title:           d.Title,
		})
	}
}

func (s *Service) insertPQRS(tx *gorm.DB, assessmentEyeID int64, items []PQRSInput) {
	for _, p := range items {
		tx.Create(&assessmentModel.AssessmentPQRS{
			AssessmentEyeID: assessmentEyeID,
			PqrsID:          p.IDPqrs,
		})
	}
}

// ---------- SaveAssessment ----------

func (s *Service) SaveAssessment(username string, examID int64, input SaveAssessmentInput) ([]AssessmentResult, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}
	if err := s.validateExam(emp, examID); err != nil {
		return nil, err
	}
	if len(input.Assessments) == 0 {
		return nil, errors.New("assessments required")
	}

	var created []assessmentModel.AssessmentEye
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, ai := range input.Assessments {
			ae := assessmentModel.AssessmentEye{
				EyeExamID:  examID,
				Impression: ai.Impression,
				Plan:       ai.Plan,
			}
			if err := tx.Create(&ae).Error; err != nil {
				return err
			}
			s.insertDiagnoses(tx, ae.IDAssessmentEye, ai.Diagnosis)
			s.insertPQRS(tx, ae.IDAssessmentEye, ai.PQRS)
			created = append(created, ae)
		}
		activitylog.Log(tx, "exam_assessment", "save", activitylog.WithEntity(examID))
		return nil
	})
	if err != nil {
		return nil, err
	}

	results := make([]AssessmentResult, 0, len(created))
	for _, ae := range created {
		results = append(results, s.buildAssessmentResult(s.db, ae))
	}
	return results, nil
}

// ---------- GetAssessments ----------

func (s *Service) GetAssessments(examID int64) ([]AssessmentResult, error) {
	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}

	var assessments []assessmentModel.AssessmentEye
	s.db.Where("eye_exam_id = ?", examID).Find(&assessments)

	results := make([]AssessmentResult, 0, len(assessments))
	for _, ae := range assessments {
		results = append(results, s.buildAssessmentResult(s.db, ae))
	}
	return results, nil
}

// ---------- UpdateAssessment ----------

func (s *Service) UpdateAssessment(username string, examID int64, input UpdateAssessmentInput) ([]AssessmentResult, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}
	if err := s.validateExam(emp, examID); err != nil {
		return nil, err
	}
	if len(input.Assessments) == 0 {
		return nil, errors.New("assessments required")
	}

	var updated []assessmentModel.AssessmentEye
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, ai := range input.Assessments {
			var ae assessmentModel.AssessmentEye
			if ai.IDAssessmentEye != nil {
				if err := tx.First(&ae, *ai.IDAssessmentEye).Error; err != nil {
					return errors.New("assessment not found")
				}
				if ae.EyeExamID != examID {
					return errors.New("assessment does not belong to exam")
				}
			} else {
				ae = assessmentModel.AssessmentEye{EyeExamID: examID}
				if err := tx.Create(&ae).Error; err != nil {
					return err
				}
			}

			if ai.Impression != nil {
				ae.Impression = ai.Impression
			}
			if ai.Plan != nil {
				ae.Plan = ai.Plan
			}
			if err := tx.Save(&ae).Error; err != nil {
				return err
			}

			tx.Where("assessment_eye_id = ?", ae.IDAssessmentEye).Delete(&assessmentModel.AssessmentDiagnosis{})
			tx.Where("assessment_eye_id = ?", ae.IDAssessmentEye).Delete(&assessmentModel.AssessmentPQRS{})
			s.insertDiagnoses(tx, ae.IDAssessmentEye, ai.Diagnosis)
			s.insertPQRS(tx, ae.IDAssessmentEye, ai.PQRS)
			updated = append(updated, ae)
		}
		activitylog.Log(tx, "exam_assessment", "update", activitylog.WithEntity(examID))
		return nil
	})
	if err != nil {
		return nil, err
	}

	results := make([]AssessmentResult, 0, len(updated))
	for _, ae := range updated {
		results = append(results, s.buildAssessmentResult(s.db, ae))
	}
	return results, nil
}

// ---------- DeleteAssessmentDiagnosis ----------

func (s *Service) DeleteAssessmentDiagnosis(username string, examID, assessmentID, diagnosisID int64) error {
	emp, err := s.getEmployee(username)
	if err != nil {
		return err
	}
	if err := s.validateExam(emp, examID); err != nil {
		return err
	}

	var ae assessmentModel.AssessmentEye
	if err := s.db.Where("id_assessment_eye = ? AND eye_exam_id = ?", assessmentID, examID).First(&ae).Error; err != nil {
		return errors.New("assessment not found")
	}

	var diag assessmentModel.AssessmentDiagnosis
	if err := s.db.Where("id_assessment_diagnosis = ? AND assessment_eye_id = ?", diagnosisID, assessmentID).First(&diag).Error; err != nil {
		return errors.New("diagnosis not found")
	}

	if err := s.db.Delete(&diag).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "exam_assessment", "delete_diagnosis", activitylog.WithEntity(examID))
	return nil
}

// ---------- DeleteAssessmentPQRS ----------

// pqrsID here is the pqrs_id FK (id_pqrs), not id_assessment_pqrs
func (s *Service) DeleteAssessmentPQRS(username string, examID, assessmentID, pqrsID int64) error {
	emp, err := s.getEmployee(username)
	if err != nil {
		return err
	}
	if err := s.validateExam(emp, examID); err != nil {
		return err
	}

	var ae assessmentModel.AssessmentEye
	if err := s.db.Where("id_assessment_eye = ? AND eye_exam_id = ?", assessmentID, examID).First(&ae).Error; err != nil {
		return errors.New("assessment not found")
	}

	var pqrs assessmentModel.AssessmentPQRS
	if err := s.db.Where("assessment_eye_id = ? AND pqrs_id = ?", assessmentID, pqrsID).First(&pqrs).Error; err != nil {
		return errors.New("pqrs not found")
	}

	if err := s.db.Delete(&pqrs).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "exam_assessment", "delete_pqrs", activitylog.WithEntity(examID))
	return nil
}

// ---------- DeleteAssessment ----------

func (s *Service) DeleteAssessment(username string, examID, assessmentID int64) error {
	emp, err := s.getEmployee(username)
	if err != nil {
		return err
	}
	if err := s.validateExam(emp, examID); err != nil {
		return err
	}

	var ae assessmentModel.AssessmentEye
	if err := s.db.Where("id_assessment_eye = ? AND eye_exam_id = ?", assessmentID, examID).First(&ae).Error; err != nil {
		return errors.New("assessment not found")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("assessment_eye_id = ?", assessmentID).Delete(&assessmentModel.AssessmentDiagnosis{})
		tx.Where("assessment_eye_id = ?", assessmentID).Delete(&assessmentModel.AssessmentPQRS{})
		if err := tx.Delete(&ae).Error; err != nil {
			return err
		}
		activitylog.Log(tx, "exam_assessment", "delete", activitylog.WithEntity(examID))
		return nil
	})
}

// ---------- SearchDiagnosis ----------

func (s *Service) SearchDiagnosis(term string) []SearchResult {
	pattern := term + "%"
	var results []SearchResult

	var chapters []diseaseModel.ChapterDisease
	s.db.Where("letter ILIKE ?", pattern).Limit(10).Find(&chapters)
	for _, c := range chapters {
		results = append(results, SearchResult{LevelID: c.IDChapterDisease, Code: c.Letter, Title: c.Title, Type: "chapter"})
	}

	var groups []diseaseModel.GroupDisease
	s.db.Where("code ILIKE ?", pattern).Limit(10).Find(&groups)
	for _, g := range groups {
		results = append(results, SearchResult{LevelID: g.IDGroupDisease, Code: g.Code, Title: g.Title, Type: "group"})
	}

	var subgroups []diseaseModel.SubgroupDisease
	s.db.Where("code ILIKE ?", pattern).Limit(10).Find(&subgroups)
	for _, sg := range subgroups {
		results = append(results, SearchResult{LevelID: sg.IDSubgroupDisease, Code: sg.Code, Title: sg.Title, Type: "subgroup"})
	}

	var level4s []diseaseModel.SubgroupDiseaseLevel4
	s.db.Where("code ILIKE ?", pattern).Limit(10).Find(&level4s)
	for _, l4 := range level4s {
		results = append(results, SearchResult{LevelID: l4.IDLevel4, Code: l4.Code, Title: l4.TitleLevel4, Type: "level_4"})
	}

	var diagnoses []diseaseModel.Diagnosis
	s.db.Where("code ILIKE ?", pattern).Limit(10).Find(&diagnoses)
	for _, d := range diagnoses {
		results = append(results, SearchResult{LevelID: d.IDDiagnosis, Code: d.Code, Title: d.TitleDiagnosis, Type: "diagnosis"})
	}

	var level6s []diseaseModel.DiagnosisLevel6
	s.db.Where("code ILIKE ?", pattern).Limit(10).Find(&level6s)
	for _, l6 := range level6s {
		results = append(results, SearchResult{LevelID: l6.IDLevel6, Code: l6.Code, Title: l6.TitleLevel6, Type: "level_6"})
	}

	if len(results) > 10 {
		results = results[:10]
	}
	if results == nil {
		results = []SearchResult{}
	}
	return results
}

// ---------- GetMyTopDiseases ----------

func (s *Service) GetMyTopDiseases() []TopDiseaseResult {
	var diseases []diseaseModel.MyTopDisease
	s.db.Find(&diseases)

	results := make([]TopDiseaseResult, 0, len(diseases))
	for _, d := range diseases {
		results = append(results, TopDiseaseResult{
			LevelID:  d.LevelID,
			Code:     d.Code,
			Title:    d.Title,
			Type:     d.Type,
			GroupSet: d.GroupSet,
			Children: s.loadDiseaseChildren(d.Type, d.LevelID),
		})
	}
	return results
}

func (s *Service) loadDiseaseChildren(diseaseType string, levelID int64) []DiseaseChildResult {
	children := []DiseaseChildResult{}
	switch diseaseType {
	case "chapter":
		var rows []diseaseModel.GroupDisease
		s.db.Where("chapter_disease_id_chapter_disease = ?", levelID).Limit(10).Find(&rows)
		for _, r := range rows {
			children = append(children, DiseaseChildResult{LevelID: r.IDGroupDisease, Code: r.Code, Title: r.Title, Type: "group"})
		}
	case "group":
		var rows []diseaseModel.SubgroupDisease
		s.db.Where("group_disease_id_group_disease = ?", levelID).Limit(10).Find(&rows)
		for _, r := range rows {
			children = append(children, DiseaseChildResult{LevelID: r.IDSubgroupDisease, Code: r.Code, Title: r.Title, Type: "subgroup"})
		}
	case "subgroup":
		var rows []diseaseModel.SubgroupDiseaseLevel4
		s.db.Where("subgroup_disease_id_subgroup_disease = ?", levelID).Limit(10).Find(&rows)
		for _, r := range rows {
			children = append(children, DiseaseChildResult{LevelID: r.IDLevel4, Code: r.Code, Title: r.TitleLevel4, Type: "level_4"})
		}
	case "level_4":
		var rows []diseaseModel.Diagnosis
		s.db.Where("level_4_id = ?", levelID).Limit(10).Find(&rows)
		for _, r := range rows {
			children = append(children, DiseaseChildResult{LevelID: r.IDDiagnosis, Code: r.Code, Title: r.TitleDiagnosis, Type: "diagnosis"})
		}
	case "diagnosis":
		var rows []diseaseModel.DiagnosisLevel6
		s.db.Where("diagnosis_id = ?", levelID).Limit(10).Find(&rows)
		for _, r := range rows {
			children = append(children, DiseaseChildResult{LevelID: r.IDLevel6, Code: r.Code, Title: r.TitleLevel6, Type: "level_6"})
		}
	}
	return children
}

// ---------- AddMyTopDisease ----------

func (s *Service) AddMyTopDisease(username string, input AddTopDiseaseInput) (*diseaseModel.MyTopDisease, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	if !s.diseaseExists(input.Type, input.LevelID) {
		return nil, errors.New("disease not found")
	}

	var existing diseaseModel.MyTopDisease
	if err := s.db.Where("level_id = ? AND employee_id = ?", input.LevelID, emp.IDEmployee).First(&existing).Error; err == nil {
		return nil, errors.New("already in top diseases")
	}

	empID := int64(emp.IDEmployee)
	record := diseaseModel.MyTopDisease{
		LevelID:    input.LevelID,
		Type:       input.Type,
		Code:       input.Code,
		Title:      input.Title,
		GroupSet:   input.GroupSet,
		EmployeeID: &empID,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *Service) diseaseExists(diseaseType string, levelID int64) bool {
	var count int64
	switch diseaseType {
	case "chapter":
		s.db.Model(&diseaseModel.ChapterDisease{}).Where("id_chapter_disease = ?", levelID).Count(&count)
	case "group":
		s.db.Model(&diseaseModel.GroupDisease{}).Where("id_group_disease = ?", levelID).Count(&count)
	case "subgroup":
		s.db.Model(&diseaseModel.SubgroupDisease{}).Where("id_subgroup_disease = ?", levelID).Count(&count)
	case "level_4":
		s.db.Model(&diseaseModel.SubgroupDiseaseLevel4{}).Where("id_level_4 = ?", levelID).Count(&count)
	case "diagnosis":
		s.db.Model(&diseaseModel.Diagnosis{}).Where("id_diagnosis = ?", levelID).Count(&count)
	case "level_6":
		s.db.Model(&diseaseModel.DiagnosisLevel6{}).Where("id_level_6 = ?", levelID).Count(&count)
	default:
		return false
	}
	return count > 0
}

// ---------- DeleteMyTopDisease ----------

func (s *Service) DeleteMyTopDisease(username string, diseaseID int64) error {
	emp, err := s.getEmployee(username)
	if err != nil {
		return err
	}

	var disease diseaseModel.MyTopDisease
	if err := s.db.Where("id_my_top_disease = ? AND employee_id = ?", diseaseID, emp.IDEmployee).First(&disease).Error; err != nil {
		return errors.New("disease not found or not authorized")
	}
	return s.db.Delete(&disease).Error
}

// ---------- GetAllPQRS ----------

func (s *Service) GetAllPQRS() []PQRSListResult {
	var pqrsList []generalModel.PQRS
	s.db.Preload("GroupPQRSRef").Find(&pqrsList)

	results := make([]PQRSListResult, 0, len(pqrsList))
	for _, p := range pqrsList {
		var group *string
		if p.GroupPQRSRef != nil {
			g := p.GroupPQRSRef.Title
			group = &g
		}
		results = append(results, PQRSListResult{
			IDPqrs: p.IDPQRS,
			Code:   p.Code,
			Title:  p.Title,
			Group:  group,
		})
	}
	return results
}
