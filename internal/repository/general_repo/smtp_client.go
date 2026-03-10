package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type SmtpClientRepo struct{ DB *gorm.DB }

func NewSmtpClientRepo(db *gorm.DB) *SmtpClientRepo { return &SmtpClientRepo{DB: db} }

func (r *SmtpClientRepo) GetActive() (*general.SmtpClient, error) {
	var v general.SmtpClient
	if err := r.DB.Where("active = true").First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *SmtpClientRepo) GetByID(id int) (*general.SmtpClient, error) {
	var v general.SmtpClient
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *SmtpClientRepo) GetAll() ([]general.SmtpClient, error) {
	var items []general.SmtpClient
	return items, r.DB.Find(&items).Error
}

func (r *SmtpClientRepo) Create(v *general.SmtpClient) error { return r.DB.Create(v).Error }
func (r *SmtpClientRepo) Save(v *general.SmtpClient) error   { return r.DB.Save(v).Error }
func (r *SmtpClientRepo) Delete(id int) error {
	return r.DB.Delete(&general.SmtpClient{}, id).Error
}
