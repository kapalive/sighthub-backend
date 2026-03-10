package devices_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/devices"
)

type PrinterRepo struct{ DB *gorm.DB }

func NewPrinterRepo(db *gorm.DB) *PrinterRepo {
	return &PrinterRepo{DB: db}
}

func (r *PrinterRepo) GetByID(id int) (*devices.Printer, error) {
	var v devices.Printer
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *PrinterRepo) GetByLocationID(locationID int) ([]devices.Printer, error) {
	var items []devices.Printer
	return items, r.DB.Where("location_id = ?", locationID).Find(&items).Error
}

func (r *PrinterRepo) GetByDeviceID(deviceID string) (*devices.Printer, error) {
	var v devices.Printer
	if err := r.DB.Where("id_device = ?", deviceID).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *PrinterRepo) Create(v *devices.Printer) error {
	return r.DB.Create(v).Error
}

func (r *PrinterRepo) Save(v *devices.Printer) error {
	return r.DB.Save(v).Error
}

func (r *PrinterRepo) Delete(id int) error {
	return r.DB.Delete(&devices.Printer{}, id).Error
}
