package misc_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/misc"
)

type MiscInvoiceItemRepo struct{ DB *gorm.DB }

func NewMiscInvoiceItemRepo(db *gorm.DB) *MiscInvoiceItemRepo {
	return &MiscInvoiceItemRepo{DB: db}
}

func (r *MiscInvoiceItemRepo) GetByID(id int64) (*misc.MiscInvoiceItem, error) {
	var item misc.MiscInvoiceItem
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *MiscInvoiceItemRepo) GetByItemNumber(itemNumber string) (*misc.MiscInvoiceItem, error) {
	var item misc.MiscInvoiceItem
	if err := r.DB.Where("item_number = ?", itemNumber).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *MiscInvoiceItemRepo) GetAll() ([]misc.MiscInvoiceItem, error) {
	var items []misc.MiscInvoiceItem
	if err := r.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MiscInvoiceItemRepo) GetActive() ([]misc.MiscInvoiceItem, error) {
	var items []misc.MiscInvoiceItem
	if err := r.DB.Where("active = true").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MiscInvoiceItemRepo) Search(query string) ([]misc.MiscInvoiceItem, error) {
	var items []misc.MiscInvoiceItem
	q := "%" + query + "%"
	if err := r.DB.Where("item_number ILIKE ? OR description ILIKE ?", q, q).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MiscInvoiceItemRepo) Create(item *misc.MiscInvoiceItem) error {
	return r.DB.Create(item).Error
}

func (r *MiscInvoiceItemRepo) Save(item *misc.MiscInvoiceItem) error {
	return r.DB.Save(item).Error
}

func (r *MiscInvoiceItemRepo) Delete(id int64) error {
	return r.DB.Delete(&misc.MiscInvoiceItem{}, id).Error
}
