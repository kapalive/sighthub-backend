// internal/repository/lab_ticket_repo/lab_ticket.go
package lab_ticket_repo

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	lt "sighthub-backend/internal/models/lab_ticket"
	"sighthub-backend/internal/models/vendors"
)

type LabTicketRepo struct{ DB *gorm.DB }

func NewLabTicketRepo(db *gorm.DB) *LabTicketRepo { return &LabTicketRepo{DB: db} }

// ─────────────────────────────────────────────
// LIST / SEARCH
// ─────────────────────────────────────────────

// LabTicketListItem — краткое представление для списков.
type LabTicketListItem struct {
	IDLabTicket       int64      `json:"id_lab_ticket"`
	NumberTicket      string     `json:"number_ticket"`
	GOrC              *string    `json:"g_or_c,omitempty"`
	DateCreate        *time.Time `json:"date_create,omitempty"`
	DatePromise       *time.Time `json:"date_promise,omitempty"`
	PatientID         int64      `json:"patient_id"`
	InvoiceID         int64      `json:"invoice_id"`
	LabTicketStatusID int64      `json:"lab_ticket_status_id"`
	StatusName        string     `json:"status_name"`
	LabID             *int       `json:"lab_id,omitempty"`
	LabName           *string    `json:"lab_name,omitempty"`
}

// LabTicketFilterInput — параметры фильтрации списка.
type LabTicketFilterInput struct {
	LocationID *int64
	DateFrom   *time.Time
	DateTo     *time.Time
	StatusID   *int64
}

// GetList возвращает тикеты с фильтрами.
func (r *LabTicketRepo) GetList(f LabTicketFilterInput) ([]LabTicketListItem, error) {
	q := r.DB.
		Table("lab_ticket t").
		Select("t.id_lab_ticket, t.number_ticket, t.g_or_c, t.date_create, t.date_promise, t.patient_id, t.invoice_id, t.lab_ticket_status_id, s.ticket_status AS status_name, t.lab_id, l.title_lab AS lab_name").
		Joins("LEFT JOIN lab_ticket_status s ON s.id_lab_ticket_status = t.lab_ticket_status_id").
		Joins("LEFT JOIN lab l ON l.id_lab = t.lab_id")

	if f.LocationID != nil {
		q = q.Joins("JOIN invoice i ON i.id_invoice = t.invoice_id").
			Where("i.location_id = ?", *f.LocationID)
	}
	if f.StatusID != nil {
		q = q.Where("t.lab_ticket_status_id = ?", *f.StatusID)
	}
	if f.DateFrom != nil {
		q = q.Where("t.date_create >= ?", f.DateFrom.Format("2006-01-02"))
	}
	if f.DateTo != nil {
		q = q.Where("t.date_create <= ?", f.DateTo.Format("2006-01-02"))
	}

	var rows []LabTicketListItem
	return rows, q.Order("t.date_create DESC").Scan(&rows).Error
}

// GetByInvoiceID возвращает все тикеты для инвойса.
func (r *LabTicketRepo) GetByInvoiceID(invoiceID int64) ([]lt.LabTicket, error) {
	var rows []lt.LabTicket
	return rows, r.DB.
		Preload("LabTicketStatus").
		Preload("Lab").
		Where("invoice_id = ?", invoiceID).
		Find(&rows).Error
}

// GetByID возвращает полный тикет со всеми sub-записями.
func (r *LabTicketRepo) GetByID(id int64) (*lt.LabTicket, error) {
	var row lt.LabTicket
	err := r.DB.
		Preload("LabTicketStatus").
		Preload("Lab").
		Preload("Powers").
		Preload("Lens").
		Preload("Lens.LensType").
		Preload("Lens.LensesMaterial").
		Preload("Lens.LensEdge").
		Preload("Lens.LensTintColor").
		Preload("Frame").
		Preload("Frame.FrameTypeMaterial").
		Preload("Frame.FrameShape").
		Preload("PowersContact").
		Preload("Contact").
		Preload("Contact.ContactLensService").
		Preload("Contact.BrandContactLens").
		Preload("Contact.Manufacturer").
		First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Search ищет тикет по номеру тикета или номеру инвойса.
func (r *LabTicketRepo) Search(query string, locationID *int64) ([]LabTicketListItem, error) {
	q := r.DB.
		Table("lab_ticket t").
		Select("t.id_lab_ticket, t.number_ticket, t.g_or_c, t.date_create, t.date_promise, t.patient_id, t.invoice_id, t.lab_ticket_status_id, s.ticket_status AS status_name, t.lab_id, l.title_lab AS lab_name").
		Joins("LEFT JOIN lab_ticket_status s ON s.id_lab_ticket_status = t.lab_ticket_status_id").
		Joins("LEFT JOIN lab l ON l.id_lab = t.lab_id").
		Joins("JOIN invoice i ON i.id_invoice = t.invoice_id").
		Where("t.number_ticket ILIKE ? OR i.number_invoice ILIKE ?", "%"+query+"%", "%"+query+"%")
	if locationID != nil {
		q = q.Where("i.location_id = ?", *locationID)
	}
	var rows []LabTicketListItem
	return rows, q.Order("t.date_create DESC").Limit(20).Scan(&rows).Error
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// CreateLabTicketInput — данные для создания тикета.
type CreateLabTicketInput struct {
	NumberTicket      string
	GOrC              *string // "g" | "c"
	DateCreate        *time.Time
	DatePromise       *time.Time
	ShipTo            *string
	LabID             *int
	LabTicketStatusID int64
	PatientID         int64
	OrdersLensID      *int64
	InvoiceID         int64
	Tray              *string
	OurNote           *string
	LabInstructions   *string
	EmployeeID        int64
}

// Create создаёт базовый тикет.
func (r *LabTicketRepo) Create(inp CreateLabTicketInput) (*lt.LabTicket, error) {
	ticket := &lt.LabTicket{
		NumberTicket:      inp.NumberTicket,
		GOrC:              inp.GOrC,
		DateCreate:        inp.DateCreate,
		DatePromise:       inp.DatePromise,
		ShipTo:            inp.ShipTo,
		LabID:             inp.LabID,
		LabTicketStatusID: inp.LabTicketStatusID,
		PatientID:         inp.PatientID,
		OrdersLensID:      inp.OrdersLensID,
		InvoiceID:         inp.InvoiceID,
		Tray:              inp.Tray,
		OurNote:           inp.OurNote,
		LabInstructions:   inp.LabInstructions,
		EmployeeID:        inp.EmployeeID,
	}
	return ticket, r.DB.Create(ticket).Error
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

// UpdateLabTicketInput — изменяемые поля тикета.
type UpdateLabTicketInput struct {
	GOrC                        *string
	DateCreate                  *time.Time
	DatePromise                 *time.Time
	ShipTo                      *string
	LabID                       *int
	LabTicketStatusID           *int64
	OrdersLensID                *int64
	Tray                        *string
	Notified                    *string
	Amt                         *string
	LabTicketPowersID           *int64
	LabTicketLensID             *int64
	LabTicketFrameID            *int64
	LabTicketPowersContactID    *int64
	LabTicketContactID          *int64
	OurNote                     *string
	LabInstructions             *string
}

// Update обновляет поля тикета.
func (r *LabTicketRepo) Update(id int64, inp UpdateLabTicketInput) error {
	updates := map[string]interface{}{}
	if inp.GOrC != nil                     { updates["g_or_c"]                       = *inp.GOrC }
	if inp.DateCreate != nil               { updates["date_create"]                  = inp.DateCreate.Format("2006-01-02") }
	if inp.DatePromise != nil              { updates["date_promise"]                 = inp.DatePromise.Format("2006-01-02") }
	if inp.ShipTo != nil                   { updates["ship_to"]                      = *inp.ShipTo }
	if inp.LabID != nil                    { updates["lab_id"]                       = *inp.LabID }
	if inp.LabTicketStatusID != nil        { updates["lab_ticket_status_id"]         = *inp.LabTicketStatusID }
	if inp.OrdersLensID != nil             { updates["orders_lens_id"]               = *inp.OrdersLensID }
	if inp.Tray != nil                     { updates["tray"]                         = *inp.Tray }
	if inp.Notified != nil                 { updates["notified"]                     = *inp.Notified }
	if inp.Amt != nil                      { updates["amt"]                          = *inp.Amt }
	if inp.LabTicketPowersID != nil        { updates["lab_ticket_powers_id"]         = *inp.LabTicketPowersID }
	if inp.LabTicketLensID != nil          { updates["lab_ticket_lens_id"]           = *inp.LabTicketLensID }
	if inp.LabTicketFrameID != nil         { updates["lab_ticket_frame_id"]          = *inp.LabTicketFrameID }
	if inp.LabTicketPowersContactID != nil { updates["lab_ticket_powers_contact_id"] = *inp.LabTicketPowersContactID }
	if inp.LabTicketContactID != nil       { updates["lab_ticket_contact_id"]        = *inp.LabTicketContactID }
	if inp.OurNote != nil                  { updates["our_note"]                     = *inp.OurNote }
	if inp.LabInstructions != nil          { updates["lab_instructions"]             = *inp.LabInstructions }
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&lt.LabTicket{}).Where("id_lab_ticket = ?", id).Updates(updates).Error
}

// UpdateStatus обновляет только статус тикета.
func (r *LabTicketRepo) UpdateStatus(id, statusID int64) error {
	return r.DB.Model(&lt.LabTicket{}).
		Where("id_lab_ticket = ?", id).
		Update("lab_ticket_status_id", statusID).Error
}

// ─────────────────────────────────────────────
// REFERENCE: Labs
// ─────────────────────────────────────────────

// GetLabs возвращает список лабораторий (vendors with lab=true).
func (r *LabTicketRepo) GetLabs() ([]vendors.Vendor, error) {
	var rows []vendors.Vendor
	return rows, r.DB.Where("lab = true").Order("vendor_name").Find(&rows).Error
}

// ─────────────────────────────────────────────
// NUMBER GENERATION
// ─────────────────────────────────────────────

// NextTicketNumber генерирует следующий номер тикета.
// Формат: T<4-digit-sequential> (напр. T0042)
func (r *LabTicketRepo) NextTicketNumber() (string, error) {
	var maxNum string
	r.DB.Model(&lt.LabTicket{}).
		Select("COALESCE(MAX(number_ticket), 'T0000')::text").
		Scan(&maxNum)
	var n int
	fmt.Sscanf(maxNum[1:], "%d", &n)
	return fmt.Sprintf("T%04d", n+1), nil
}

func (r *LabTicketRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
