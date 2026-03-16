// internal/repository/invoices_repo/invoice.go
package invoices_repo

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/invoices"
)

type InvoiceRepo struct{ DB *gorm.DB }

func NewInvoiceRepo(db *gorm.DB) *InvoiceRepo { return &InvoiceRepo{DB: db} }

// ─────────────────────────────────────────────
// LIST / SEARCH
// ─────────────────────────────────────────────

// InvoiceListItem — краткое представление инвойса для списков.
type InvoiceListItem struct {
	IDInvoice       int64     `json:"id_invoice"`
	NumberInvoice   string    `json:"number_invoice"`
	DateCreate      time.Time `json:"date_create"`
	TotalAmount     float64   `json:"total_amount"`
	FinalAmount     float64   `json:"final_amount"`
	Due             float64   `json:"due"`
	StatusInvoiceID int64     `json:"status_invoice_id"`
	StatusName      string    `json:"status_name"`
	PatientID       int64     `json:"patient_id"`
	LocationID      int64     `json:"location_id"`
	Finalized       bool      `json:"finalized"`
}

// GetForPatient возвращает инвойсы пациента в локации (только patient-инвойсы, т.е. number НЕ начинается с 'V').
func (r *InvoiceRepo) GetForPatient(patientID, locationID int64) ([]InvoiceListItem, error) {
	var rows []InvoiceListItem
	err := r.DB.
		Table("invoice i").
		Select("i.id_invoice, i.number_invoice, i.date_create, i.total_amount, i.final_amount, i.due, i.status_invoice_id, s.status_invoice_value AS status_name, i.patient_id, i.location_id, i.finalized").
		Joins("LEFT JOIN status_invoice s ON s.id_status_invoice = i.status_invoice_id").
		Where("i.patient_id = ? AND i.location_id = ? AND i.number_invoice NOT LIKE 'V%'", patientID, locationID).
		Order("i.date_create DESC").
		Scan(&rows).Error
	return rows, err
}

// GetByID возвращает инвойс по ID (без preload).
func (r *InvoiceRepo) GetByID(id int64) (*invoices.Invoice, error) {
	var row invoices.Invoice
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetDetail возвращает инвойс с предзагруженными связями.
func (r *InvoiceRepo) GetDetail(id int64) (*invoices.Invoice, error) {
	var row invoices.Invoice
	err := r.DB.
		Preload("StatusInvoice").
		Preload("PaymentMethod").
		First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Search ищет инвойсы по частичному совпадению номера (до 20 результатов).
func (r *InvoiceRepo) Search(query string, locationID int64) ([]InvoiceListItem, error) {
	var rows []InvoiceListItem
	err := r.DB.
		Table("invoice i").
		Select("i.id_invoice, i.number_invoice, i.date_create, i.total_amount, i.final_amount, i.due, i.status_invoice_id, s.status_invoice_value AS status_name, i.patient_id, i.location_id, i.finalized").
		Joins("LEFT JOIN status_invoice s ON s.id_status_invoice = i.status_invoice_id").
		Where("i.location_id = ? AND i.number_invoice ILIKE ?", locationID, "%"+query+"%").
		Order("i.date_create DESC").
		Limit(20).
		Scan(&rows).Error
	return rows, err
}

// GetNonPatientInvoices возвращает инвойсы НЕ связанные с реальным пациентом
// (для Transfer/Vendor flows — number начинается с 'I' или 'V').
func (r *InvoiceRepo) GetNonPatientInvoices(locationID int64, prefix string) ([]InvoiceListItem, error) {
	var rows []InvoiceListItem
	err := r.DB.
		Table("invoice i").
		Select("i.id_invoice, i.number_invoice, i.date_create, i.total_amount, i.final_amount, i.due, i.status_invoice_id, s.status_invoice_value AS status_name, i.patient_id, i.location_id, i.finalized").
		Joins("LEFT JOIN status_invoice s ON s.id_status_invoice = i.status_invoice_id").
		Where("i.location_id = ? AND i.number_invoice LIKE ?", locationID, prefix+"%").
		Order("i.date_create DESC").
		Scan(&rows).Error
	return rows, err
}

// GetTransfers возвращает трансферные инвойсы между локациями.
type TransferRow struct {
	IDInvoice     int64     `json:"id_invoice"`
	NumberInvoice string    `json:"number_invoice"`
	DateCreate    time.Time `json:"date_create"`
	TotalAmount   float64   `json:"total_amount"`
	FinalAmount   float64   `json:"final_amount"`
	Due           float64   `json:"due"`
	LocationID    int64     `json:"location_id"`
	ToLocationID  *int64    `json:"to_location_id"`
	StatusName    string    `json:"status_name"`
	Finalized     bool      `json:"finalized"`
}

func (r *InvoiceRepo) GetTransfers(locationID int64) ([]TransferRow, error) {
	var rows []TransferRow
	err := r.DB.
		Table("invoice i").
		Select("i.id_invoice, i.number_invoice, i.date_create, i.total_amount, i.final_amount, i.due, i.location_id, i.to_location_id, s.status_invoice_value AS status_name, i.finalized").
		Joins("LEFT JOIN status_invoice s ON s.id_status_invoice = i.status_invoice_id").
		Where("(i.location_id = ? OR i.to_location_id = ?) AND i.to_location_id IS NOT NULL", locationID, locationID).
		Order("i.date_create DESC").
		Scan(&rows).Error
	return rows, err
}

// GetTransferLocations возвращает список доступных локаций для трансфера
// (все активные локации кроме текущей).
type LocationRef struct {
	IDLocation int64  `json:"id_location"`
	Name       string `json:"name"`
	StoreID    *int64 `json:"store_id,omitempty"`
}

func (r *InvoiceRepo) GetTransferLocations(currentLocationID int64) ([]LocationRef, error) {
	var rows []LocationRef
	err := r.DB.
		Table("location").
		Select("id_location, COALESCE(full_name, short_name, CAST(id_location AS VARCHAR)) AS name, store_id").
		Where("id_location <> ? AND is_active = true", currentLocationID).
		Scan(&rows).Error
	return rows, err
}

// ─────────────────────────────────────────────
// CREATE / UPDATE / DELETE
// ─────────────────────────────────────────────

// CreateInvoiceInput содержит данные для нового инвойса.
type CreateInvoiceInput struct {
	NumberInvoice   string
	PatientID       int64
	LocationID      int64
	EmployeeID      *int64
	DoctorID        *int64
	VendorID        int64  // 0 для patient-инвойсов (будет проставлен дефолтный)
	StatusInvoiceID int64
	ToLocationID    *int64
	Remake          bool
}

// Create создаёт новый инвойс.
func (r *InvoiceRepo) Create(inp CreateInvoiceInput) (*invoices.Invoice, error) {
	inv := &invoices.Invoice{
		NumberInvoice:   inp.NumberInvoice,
		PatientID:       inp.PatientID,
		LocationID:      inp.LocationID,
		EmployeeID:      inp.EmployeeID,
		DoctorID:        inp.DoctorID,
		VendorID:        &inp.VendorID,
		StatusInvoiceID: &inp.StatusInvoiceID,
		ToLocationID:    inp.ToLocationID,
		Remake:          inp.Remake,
		DateCreate:      time.Now(),
	}
	return inv, r.DB.Create(inv).Error
}

// UpdateInvoiceInput — изменяемые поля инвойса.
type UpdateInvoiceInput struct {
	PaymentMethodID     *int64
	StatusInvoiceID     *int64
	EmployeeID          *int64
	DoctorID            *int64
	InsurancePolicyID   *int64
	Discount            *float64
	TotalAmount         *float64
	FinalAmount         *float64
	Due                 *float64
	PTBal               *float64
	InsBal              *float64
	GiftCardBal         *float64
	TaxAmount           *float64
	Quantity            *int
	Notified            *string
	Referral            *string
	ClassField          *string
	Reason              *string
}

// Update обновляет поля инвойса.
func (r *InvoiceRepo) Update(id int64, inp UpdateInvoiceInput) error {
	updates := map[string]interface{}{}
	if inp.PaymentMethodID != nil  { updates["payment_method_id"]   = *inp.PaymentMethodID }
	if inp.StatusInvoiceID != nil  { updates["status_invoice_id"]   = *inp.StatusInvoiceID }
	if inp.EmployeeID != nil       { updates["employee_id"]          = *inp.EmployeeID }
	if inp.DoctorID != nil         { updates["doctor_id"]            = *inp.DoctorID }
	if inp.InsurancePolicyID != nil{ updates["insurance_policy_id"]  = *inp.InsurancePolicyID }
	if inp.Discount != nil         { updates["discount"]             = *inp.Discount }
	if inp.TotalAmount != nil      { updates["total_amount"]         = *inp.TotalAmount }
	if inp.FinalAmount != nil      { updates["final_amount"]         = *inp.FinalAmount }
	if inp.Due != nil              { updates["due"]                  = *inp.Due }
	if inp.PTBal != nil            { updates["pt_bal"]               = *inp.PTBal }
	if inp.InsBal != nil           { updates["ins_bal"]              = *inp.InsBal }
	if inp.GiftCardBal != nil      { updates["gift_card_bal"]        = *inp.GiftCardBal }
	if inp.TaxAmount != nil        { updates["tax_amount"]           = *inp.TaxAmount }
	if inp.Quantity != nil         { updates["quantity"]             = *inp.Quantity }
	if inp.Notified != nil         { updates["notified"]             = *inp.Notified }
	if inp.Referral != nil         { updates["referral"]             = *inp.Referral }
	if inp.ClassField != nil       { updates["class"]                = *inp.ClassField }
	if inp.Reason != nil           { updates["reason"]               = *inp.Reason }
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&invoices.Invoice{}).Where("id_invoice = ?", id).Updates(updates).Error
}

// DeleteEmpty удаляет инвойс только если он пустой (нет позиций, нулевые суммы).
func (r *InvoiceRepo) DeleteEmpty(id int64) error {
	var inv invoices.Invoice
	if err := r.DB.First(&inv, id).Error; err != nil {
		return err
	}
	// проверяем что инвойс действительно пустой
	var count int64
	r.DB.Model(&invoices.InvoiceItemSale{}).Where("invoice_id = ?", id).Count(&count)
	if count > 0 || inv.TotalAmount != 0 || inv.FinalAmount != 0 {
		return fmt.Errorf("invoice %d is not empty", id)
	}
	return r.DB.Delete(&invoices.Invoice{}, id).Error
}

// Delete удаляет инвойс без проверки (с каскадным удалением позиций).
func (r *InvoiceRepo) Delete(id int64) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("invoice_id = ?", id).Delete(&invoices.InvoiceItemSale{})
		tx.Where("invoice_id = ?", id).Delete(&invoices.InvoiceServicesItem{})
		return tx.Delete(&invoices.Invoice{}, id).Error
	})
}

// ─────────────────────────────────────────────
// FINALIZE / UNFINALIZE / REMAKE
// ─────────────────────────────────────────────

// Finalize блокирует инвойс для изменений.
func (r *InvoiceRepo) Finalize(id int64) error {
	return r.DB.Model(&invoices.Invoice{}).
		Where("id_invoice = ?", id).
		Update("finalized", true).Error
}

// Unfinalize разблокирует инвойс.
func (r *InvoiceRepo) Unfinalize(id int64) error {
	return r.DB.Model(&invoices.Invoice{}).
		Where("id_invoice = ?", id).
		Update("finalized", false).Error
}

// Remake создаёт копию инвойса с флагом remake=true и новым номером.
func (r *InvoiceRepo) Remake(sourceID int64, newNumber string) (*invoices.Invoice, error) {
	src, err := r.GetByID(sourceID)
	if err != nil || src == nil {
		return nil, fmt.Errorf("source invoice not found: %w", err)
	}
	newInv := &invoices.Invoice{
		NumberInvoice:   newNumber,
		PatientID:       src.PatientID,
		LocationID:      src.LocationID,
		EmployeeID:      src.EmployeeID,
		DoctorID:        src.DoctorID,
		VendorID:        src.VendorID,
		StatusInvoiceID: src.StatusInvoiceID,
		Remake:          true,
		DateCreate:      time.Now(),
	}
	return newInv, r.DB.Create(newInv).Error
}

// ─────────────────────────────────────────────
// TRANSFER / RECEIPT
// ─────────────────────────────────────────────

// MarkTransferPaid помечает трансферный инвойс как оплаченный
// обновляя статус и сбрасывая задолженность.
func (r *InvoiceRepo) MarkTransferPaid(id int64, paidStatusID int64) error {
	return r.DB.Model(&invoices.Invoice{}).
		Where("id_invoice = ?", id).
		Updates(map[string]interface{}{
			"status_invoice_id": paidStatusID,
			"due":               0,
		}).Error
}

// ─────────────────────────────────────────────
// DISCOUNT / GIFT CARD
// ─────────────────────────────────────────────

// ApplyDiscount применяет скидку к инвойсу и пересчитывает final_amount / due.
func (r *InvoiceRepo) ApplyDiscount(tx *gorm.DB, id int64, discount float64) error {
	db := r.DB
	if tx != nil {
		db = tx
	}
	var inv invoices.Invoice
	if err := db.First(&inv, id).Error; err != nil {
		return err
	}
	newFinal := inv.TotalAmount - discount + inv.TaxAmount
	newDue := newFinal - (inv.TotalAmount - inv.Due) // сохраняем уже оплаченное
	return db.Model(&invoices.Invoice{}).Where("id_invoice = ?", id).
		Updates(map[string]interface{}{
			"discount":     discount,
			"final_amount": newFinal,
			"due":          newDue,
		}).Error
}

// SetGiftCardBalance устанавливает gift_card_bal в инвойсе и пересчитывает due.
func (r *InvoiceRepo) SetGiftCardBalance(tx *gorm.DB, id int64, giftAmount float64) error {
	db := r.DB
	if tx != nil {
		db = tx
	}
	return db.Model(&invoices.Invoice{}).Where("id_invoice = ?", id).
		Updates(map[string]interface{}{
			"gift_card_bal": giftAmount,
		}).Error
}

// ─────────────────────────────────────────────
// NUMBER GENERATION HELPERS
// ─────────────────────────────────────────────

// NextPatientInvoiceNumber генерирует следующий номер инвойса для пациента.
// Формат: <prefix><6-digit-sequential>  (напр. P000123)
func (r *InvoiceRepo) NextPatientInvoiceNumber(prefix string) (string, error) {
	var maxNum string
	r.DB.Model(&invoices.Invoice{}).
		Where("number_invoice LIKE ?", prefix+"%").
		Select("COALESCE(MAX(number_invoice), ?)::text", prefix+"000000").
		Scan(&maxNum)

	// извлекаем числовую часть
	numPart := strings.TrimPrefix(maxNum, prefix)
	var n int
	fmt.Sscanf(numPart, "%d", &n)
	return fmt.Sprintf("%s%06d", prefix, n+1), nil
}

// ─────────────────────────────────────────────
// UTILITY
// ─────────────────────────────────────────────

func (r *InvoiceRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
