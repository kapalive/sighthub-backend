// internal/repository/vendors_repo/vendor_ap_invoice.go
package vendors_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type VendorAPInvoiceRepo struct{ DB *gorm.DB }

func NewVendorAPInvoiceRepo(db *gorm.DB) *VendorAPInvoiceRepo {
	return &VendorAPInvoiceRepo{DB: db}
}

// APInvoiceListItem — краткое представление для списков.
type APInvoiceListItem struct {
	IDVendorAPInvoice int64  `json:"id_vendor_ap_invoice"`
	VendorID          int    `json:"vendor_id"`
	VendorName        string `json:"vendor_name"`
	LocationID        int64  `json:"location_id"`
	InvoiceNumber     string `json:"invoice_number"`
	InvoiceDate       string `json:"invoice_date"`
	BillDueDate       string `json:"bill_due_date"`
	InvoiceAmount     string `json:"invoice_amount"`
	OpenBalance       string `json:"open_balance"`
	Status            string `json:"status"`
}

// GetList возвращает список vendor AP инвойсов для локации.
func (r *VendorAPInvoiceRepo) GetList(locationID int64) ([]APInvoiceListItem, error) {
	var rows []APInvoiceListItem
	err := r.DB.
		Table("vendor_ap_invoice a").
		Select("a.id_vendor_ap_invoice, a.vendor_id, v.vendor_name, a.location_id, a.invoice_number, TO_CHAR(a.invoice_date,'YYYY-MM-DD') AS invoice_date, TO_CHAR(a.bill_due_date,'YYYY-MM-DD') AS bill_due_date, a.invoice_amount, a.open_balance, a.status").
		Joins("JOIN vendor v ON v.id_vendor = a.vendor_id").
		Where("a.location_id = ?", locationID).
		Order("a.invoice_date DESC").
		Scan(&rows).Error
	return rows, err
}

// GetByID возвращает vendor AP инвойс с позициями.
type APInvoiceDetail struct {
	vendors.VendorAPInvoice
	Items []vendors.VendorAPInvoiceItem `json:"items"`
}

func (r *VendorAPInvoiceRepo) GetByID(id int64) (*APInvoiceDetail, error) {
	var inv vendors.VendorAPInvoice
	if err := r.DB.First(&inv, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var items []vendors.VendorAPInvoiceItem
	r.DB.Where("vendor_ap_invoice_id = ?", id).Find(&items)
	return &APInvoiceDetail{VendorAPInvoice: inv, Items: items}, nil
}

// CreateAPInvoiceInput — входные данные для создания vendor AP инвойса.
type CreateAPInvoiceInput struct {
	VendorID                int
	LocationID              int64
	EmployeeID              int64
	InvoiceNumber           string
	InvoiceDate             time.Time
	BillDueDate             time.Time
	InvoiceAmount           string
	OpenBalance             string
	TaxTotal                string
	AttachmentURL           *string
	Note                    *string
	Terms                   *int
	VendorLocationAccountID *int64
	Items                   []vendors.VendorAPInvoiceItem
}

// Create создаёт vendor AP инвойс с позициями в транзакции.
func (r *VendorAPInvoiceRepo) Create(inp CreateAPInvoiceInput) (*vendors.VendorAPInvoice, error) {
	inv := &vendors.VendorAPInvoice{
		VendorID:                inp.VendorID,
		LocationID:              inp.LocationID,
		EmployeeID:              inp.EmployeeID,
		InvoiceNumber:           inp.InvoiceNumber,
		InvoiceDate:             inp.InvoiceDate,
		BillDueDate:             inp.BillDueDate,
		InvoiceAmount:           inp.InvoiceAmount,
		OpenBalance:             inp.OpenBalance,
		TaxTotal:                inp.TaxTotal,
		AttachmentURL:           inp.AttachmentURL,
		Note:                    inp.Note,
		Terms:                   inp.Terms,
		VendorLocationAccountID: inp.VendorLocationAccountID,
		Status:                  "Open",
	}
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(inv).Error; err != nil {
			return err
		}
		for i := range inp.Items {
			inp.Items[i].VendorAPInvoiceID = inv.IDVendorAPInvoice
			if err := tx.Create(&inp.Items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return inv, err
}

// UpdateAPInvoiceInput — изменяемые поля vendor AP инвойса.
type UpdateAPInvoiceInput struct {
	InvoiceNumber           *string
	InvoiceDate             *time.Time
	BillDueDate             *time.Time
	InvoiceAmount           *string
	OpenBalance             *string
	TaxTotal                *string
	Status                  *string
	AttachmentURL           *string
	Note                    *string
	Terms                   *int
	VendorLocationAccountID *int64
}

// Update обновляет vendor AP инвойс.
func (r *VendorAPInvoiceRepo) Update(id int64, inp UpdateAPInvoiceInput) error {
	updates := map[string]interface{}{}
	if inp.InvoiceNumber != nil           { updates["invoice_number"]             = *inp.InvoiceNumber }
	if inp.InvoiceDate != nil             { updates["invoice_date"]               = *inp.InvoiceDate }
	if inp.BillDueDate != nil             { updates["bill_due_date"]              = *inp.BillDueDate }
	if inp.InvoiceAmount != nil           { updates["invoice_amount"]             = *inp.InvoiceAmount }
	if inp.OpenBalance != nil             { updates["open_balance"]               = *inp.OpenBalance }
	if inp.TaxTotal != nil                { updates["tax_total"]                  = *inp.TaxTotal }
	if inp.Status != nil                  { updates["status"]                     = *inp.Status }
	if inp.AttachmentURL != nil           { updates["attachment_url"]             = *inp.AttachmentURL }
	if inp.Note != nil                    { updates["note"]                       = *inp.Note }
	if inp.Terms != nil                   { updates["terms"]                      = *inp.Terms }
	if inp.VendorLocationAccountID != nil { updates["vendor_location_account_id"] = *inp.VendorLocationAccountID }
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&vendors.VendorAPInvoice{}).Where("id_vendor_ap_invoice = ?", id).Updates(updates).Error
}

// AddItem добавляет позицию к vendor AP инвойсу.
func (r *VendorAPInvoiceRepo) AddItem(item *vendors.VendorAPInvoiceItem) error {
	return r.DB.Create(item).Error
}

func (r *VendorAPInvoiceRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
