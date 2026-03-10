package employees_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type JobTitleRepo struct {
	DB *gorm.DB
}

func NewJobTitleRepo(db *gorm.DB) *JobTitleRepo {
	return &JobTitleRepo{DB: db}
}

// GetAll returns all job title records.
func (r *JobTitleRepo) GetAll() ([]employees.JobTitle, error) {
	var titles []employees.JobTitle
	if err := r.DB.Find(&titles).Error; err != nil {
		return nil, err
	}
	return titles, nil
}

// GetByID returns the job title with the given id.
func (r *JobTitleRepo) GetByID(id int) (*employees.JobTitle, error) {
	var title employees.JobTitle
	if err := r.DB.First(&title, "id_job_title = ?", id).Error; err != nil {
		return nil, err
	}
	return &title, nil
}

// IsDoctor returns true when the job title with the given id has doctor=true.
// Returns false (not an error) when the job title is not found.
func (r *JobTitleRepo) IsDoctor(jobTitleID int) (bool, error) {
	var title employees.JobTitle
	if err := r.DB.First(&title, "id_job_title = ?", jobTitleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return title.Doctor, nil
}
