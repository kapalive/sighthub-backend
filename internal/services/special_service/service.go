package special_service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	empLoginModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	specialModel "sighthub-backend/internal/models/medical/vision_exam/special"
	visionModel "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ---------- input types ----------

type FileInput struct {
	IDSpecialEyeFile *int64  `json:"id_special_eye_file"`
	FilesUploadPath  *string `json:"files_upload_path"`
	FileName         *string `json:"file_name"`
	DateRecord       *string `json:"date_record"` // "YYYY-MM-DD"
}

type SaveSpecialInput struct {
	SpecialTesting *string     `json:"special_testing"`
	Files          []FileInput `json:"files"`
}

type UpdateSpecialInput struct {
	SpecialTesting *string     `json:"special_testing"`
	Files          []FileInput `json:"files"`
}

// ---------- response types ----------

type SpecialResult struct {
	IDSpecialEyeExam int64                        `json:"id_special_eye_exam"`
	SpecialTesting   *string                      `json:"special_testing"`
	Files            []specialModel.SpecialEyeFile `json:"files"`
}

type GetSpecialResult struct {
	ExamID           int64                        `json:"exam_id"`
	Exists           bool                         `json:"exists"`
	IDSpecialEyeExam *int64                       `json:"id_special_eye_exam"`
	SpecialTesting   *string                      `json:"special_testing"`
	Files            []specialModel.SpecialEyeFile `json:"files"`
}

// ---------- helpers ----------

func strPtr(s string) *string { return &s }

func cleanFilePath(path string) string {
	const prefix = "/mnt/tank/data/"
	if strings.Contains(path, prefix) {
		return strings.TrimSpace(strings.SplitN(path, prefix, 2)[1])
	}
	return path
}

func parseDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

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

func (s *Service) loadFiles(specialEyeExamID int64) []specialModel.SpecialEyeFile {
	var files []specialModel.SpecialEyeFile
	s.db.Where("special_eye_exam_id = ?", specialEyeExamID).Find(&files)
	return files
}

// ---------- SaveSpecial ----------

func (s *Service) SaveSpecial(username string, examID int64, input SaveSpecialInput) (*SpecialResult, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized")
	}
	if exam.Passed {
		return nil, errors.New("cannot update: exam is completed")
	}

	var see specialModel.SpecialEyeExam
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Upsert SpecialEyeExam
		if err := tx.Where("eye_exam_id = ?", examID).First(&see).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			// Create new
			empty := ""
			see = specialModel.SpecialEyeExam{
				EyeExamID:      examID,
				SpecialTesting: &empty,
			}
			if err := tx.Create(&see).Error; err != nil {
				return err
			}
		}

		// Update special_testing if provided and non-empty
		if input.SpecialTesting != nil {
			v := strings.TrimSpace(*input.SpecialTesting)
			if v != "" {
				see.SpecialTesting = &v
				tx.Model(&see).Update("special_testing", v)
			}
		}

		// Create file records
		for _, fd := range input.Files {
			path := ""
			if fd.FilesUploadPath != nil {
				path = cleanFilePath(*fd.FilesUploadPath)
			}
			fileName := ""
			if fd.FileName != nil {
				fileName = *fd.FileName
			}
			dateRecord, err := parseDate(fd.DateRecord)
			if err != nil {
				return err
			}
			f := specialModel.SpecialEyeFile{
				SpecialEyeExamID: see.IDSpecialEyeExam,
				FilesUploadPath:  &path,
				FileName:         &fileName,
				DateRecord:       dateRecord,
			}
			if err := tx.Create(&f).Error; err != nil {
				return err
			}
		}

		activitylog.Log(tx, "exam_special", "save", activitylog.WithEntity(examID))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &SpecialResult{
		IDSpecialEyeExam: see.IDSpecialEyeExam,
		SpecialTesting:   see.SpecialTesting,
		Files:            s.loadFiles(see.IDSpecialEyeExam),
	}, nil
}

// ---------- GetSpecial ----------

func (s *Service) GetSpecial(examID int64) (*GetSpecialResult, error) {
	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}

	var see specialModel.SpecialEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&see).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &GetSpecialResult{
				ExamID: examID, Exists: false,
				Files: []specialModel.SpecialEyeFile{},
			}, nil
		}
		return nil, err
	}

	return &GetSpecialResult{
		ExamID:           examID,
		Exists:           true,
		IDSpecialEyeExam: &see.IDSpecialEyeExam,
		SpecialTesting:   see.SpecialTesting,
		Files:            s.loadFiles(see.IDSpecialEyeExam),
	}, nil
}

// ---------- UpdateSpecial ----------

func (s *Service) UpdateSpecial(username string, examID int64, input UpdateSpecialInput) (*SpecialResult, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized")
	}
	if exam.Passed {
		return nil, errors.New("cannot update: exam is completed")
	}

	var see specialModel.SpecialEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&see).Error; err != nil {
		return nil, errors.New("special_eye_exam not found")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update special_testing
		if input.SpecialTesting != nil {
			see.SpecialTesting = input.SpecialTesting
		}
		if err := tx.Save(&see).Error; err != nil {
			return err
		}

		// Build set of incoming file IDs (those with existing IDs)
		newFileIDs := make(map[int64]bool)
		for _, fd := range input.Files {
			if fd.IDSpecialEyeFile != nil {
				newFileIDs[*fd.IDSpecialEyeFile] = true
			}
		}

		// Delete files not present in the new list
		var existing []specialModel.SpecialEyeFile
		tx.Where("special_eye_exam_id = ?", see.IDSpecialEyeExam).Find(&existing)
		for _, f := range existing {
			if !newFileIDs[f.IDSpecialEyeFile] {
				tx.Delete(&f)
			}
		}

		// Add or update files
		for _, fd := range input.Files {
			dateRecord, err := parseDate(fd.DateRecord)
			if err != nil {
				return err
			}
			if fd.IDSpecialEyeFile != nil {
				// Update existing
				var f specialModel.SpecialEyeFile
				if err := tx.First(&f, *fd.IDSpecialEyeFile).Error; err == nil {
					if fd.FilesUploadPath != nil {
						f.FilesUploadPath = fd.FilesUploadPath
					}
					if fd.FileName != nil {
						f.FileName = fd.FileName
					}
					if fd.DateRecord != nil {
						f.DateRecord = dateRecord
					}
					tx.Save(&f)
				}
			} else {
				// New file
				path := ""
				if fd.FilesUploadPath != nil {
					path = *fd.FilesUploadPath
				}
				fileName := ""
				if fd.FileName != nil {
					fileName = *fd.FileName
				}
				f := specialModel.SpecialEyeFile{
					SpecialEyeExamID: see.IDSpecialEyeExam,
					FilesUploadPath:  &path,
					FileName:         &fileName,
					DateRecord:       dateRecord,
				}
				if err := tx.Create(&f).Error; err != nil {
					return err
				}
			}
		}

		activitylog.Log(tx, "exam_special", "update", activitylog.WithEntity(examID))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &SpecialResult{
		IDSpecialEyeExam: see.IDSpecialEyeExam,
		SpecialTesting:   see.SpecialTesting,
		Files:            s.loadFiles(see.IDSpecialEyeExam),
	}, nil
}

// ---------- DeleteSpecialEyeFile ----------

func (s *Service) DeleteSpecialEyeFile(username string, fileID int64) (int64, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return 0, err
	}

	var f specialModel.SpecialEyeFile
	if err := s.db.First(&f, fileID).Error; err != nil {
		return 0, errors.New("special_eye_file not found")
	}

	var see specialModel.SpecialEyeExam
	if err := s.db.First(&see, f.SpecialEyeExamID).Error; err != nil {
		return 0, errors.New("special_eye_exam not found")
	}

	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", see.EyeExamID).First(&exam).Error; err != nil {
		return 0, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return 0, errors.New("not authorized")
	}
	if exam.Passed {
		return 0, errors.New("cannot update: exam is completed")
	}

	if err := s.db.Delete(&f).Error; err != nil {
		return 0, err
	}
	activitylog.Log(s.db, "exam_special", "file_delete", activitylog.WithEntity(fileID))
	return fileID, nil
}
