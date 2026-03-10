package employees_repo

import (
	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type WorkShiftRepo struct {
	DB *gorm.DB
}

func NewWorkShiftRepo(db *gorm.DB) *WorkShiftRepo {
	return &WorkShiftRepo{DB: db}
}

// GetByID returns the WorkShift with the given id.
func (r *WorkShiftRepo) GetByID(id int64) (*employees.WorkShift, error) {
	var ws employees.WorkShift
	if err := r.DB.First(&ws, "id_work_shift = ?", id).Error; err != nil {
		return nil, err
	}
	return &ws, nil
}

// Create inserts a new WorkShift record and populates its primary key.
func (r *WorkShiftRepo) Create(ws *employees.WorkShift) error {
	return r.DB.Create(ws).Error
}

// Update saves all fields of the supplied WorkShift record using its primary key.
func (r *WorkShiftRepo) Update(ws *employees.WorkShift) error {
	return r.DB.Save(ws).Error
}
