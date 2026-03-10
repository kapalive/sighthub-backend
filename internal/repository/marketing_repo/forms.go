package marketing_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type FormRepo struct{ DB *gorm.DB }

func NewFormRepo(db *gorm.DB) *FormRepo { return &FormRepo{DB: db} }

func (r *FormRepo) GetAll() ([]marketing.Form, error) {
	var forms []marketing.Form
	if err := r.DB.Find(&forms).Error; err != nil {
		return nil, err
	}
	return forms, nil
}

func (r *FormRepo) GetActive() ([]marketing.Form, error) {
	var forms []marketing.Form
	if err := r.DB.Where("is_active = true").Find(&forms).Error; err != nil {
		return nil, err
	}
	return forms, nil
}

func (r *FormRepo) GetByID(id int) (*marketing.Form, error) {
	var form marketing.Form
	if err := r.DB.First(&form, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &form, nil
}

func (r *FormRepo) Create(form *marketing.Form) error {
	return r.DB.Create(form).Error
}

func (r *FormRepo) Save(form *marketing.Form) error {
	return r.DB.Save(form).Error
}

func (r *FormRepo) Delete(id int) error {
	return r.DB.Delete(&marketing.Form{}, id).Error
}
