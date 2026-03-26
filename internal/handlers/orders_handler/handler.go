package orders_handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"sighthub-backend/internal/middleware"
	"sighthub-backend/pkg/communication"
)

type Handler struct{ db *gorm.DB }

func New(db *gorm.DB) *Handler { return &Handler{db: db} }

type labStatusRow struct {
	IDLabTicket     int64      `gorm:"column:id_lab_ticket"`
	NumberTicket    string     `gorm:"column:number_ticket"`
	Tray            *string    `gorm:"column:tray"`
	InvoiceID       int64      `gorm:"column:invoice_id"`
	InvDate         *time.Time `gorm:"column:inv_date"`
	DatePromise     *time.Time `gorm:"column:date_promise"`
	DateComplete    *time.Time `gorm:"column:date_complete"`
	Late            int        `gorm:"column:late"`
	StatusID        int        `gorm:"column:lab_ticket_status_id"`
	StatusName      string     `gorm:"column:ticket_status"`
	EmployeeFirst   string     `gorm:"column:emp_first"`
	EmployeeLast    string     `gorm:"column:emp_last"`
	PatientID       *int64     `gorm:"column:patient_id"`
	PatientFirst    string     `gorm:"column:pt_first"`
	PatientLast     string     `gorm:"column:pt_last"`
	PatientPhone    *string    `gorm:"column:pt_phone"`
	Notified        *string    `gorm:"column:notified"`
	DashboardNote   *string    `gorm:"column:dashboard_note"`
	GOrC            string     `gorm:"column:g_or_c"`
	LocationID      int        `gorm:"column:location_id"`
	LocationName    string     `gorm:"column:location_name"`
}

// GET /api/orders/ticket-statuses
func (h *Handler) TicketStatuses(w http.ResponseWriter, r *http.Request) {
	type row struct {
		ID   int    `gorm:"column:id_lab_ticket_status"`
		Name string `gorm:"column:ticket_status"`
	}
	var rows []row
	h.db.Table("lab_ticket_status").Order("id_lab_ticket_status").Find(&rows)
	items := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		items = append(items, map[string]interface{}{
			"id":   r.ID,
			"name": r.Name,
		})
	}
	jsonResp(w, items, http.StatusOK)
}

// GET /api/orders/invoice-statuses
func (h *Handler) InvoiceStatuses(w http.ResponseWriter, r *http.Request) {
	statusType := r.URL.Query().Get("type")
	if statusType == "" {
		statusType = "patient"
	}
	type row struct {
		ID   int    `gorm:"column:id_status_invoice"`
		Name string `gorm:"column:status_invoice_value"`
		Type string `gorm:"column:status_type"`
	}
	var rows []row
	h.db.Table("status_invoice").Where("status_type = ?", statusType).Order("id_status_invoice").Find(&rows)
	items := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		items = append(items, map[string]interface{}{
			"id":   r.ID,
			"name": r.Name,
			"type": r.Type,
		})
	}
	jsonResp(w, items, http.StatusOK)
}

// GET /api/orders/lab-status
func (h *Handler) LabStatus(w http.ResponseWriter, r *http.Request) {
	h.orderStatus(w, r, "g")
}

// GET /api/orders/contact-status
func (h *Handler) ContactStatus(w http.ResponseWriter, r *http.Request) {
	h.orderStatus(w, r, "c")
}

func (h *Handler) orderStatus(w http.ResponseWriter, r *http.Request, gOrC string) {
	permittedIDs := middleware.PermittedLocationIDsFromCtx(r.Context())
	emp := middleware.EmployeeFromCtx(r.Context())

	// Get current location from employee
	var currentLocationID int
	if emp != nil && emp.LocationID != nil {
		currentLocationID = int(*emp.LocationID)
	}
	if currentLocationID == 0 && len(permittedIDs) > 0 {
		currentLocationID = permittedIDs[0]
	}

	q := r.URL.Query()

	// Location filter: default = current, "all" = all permitted
	locationFilter := q.Get("location_id")
	var locationIDs []int
	if locationFilter == "all" {
		locationIDs = permittedIDs
	} else if locationFilter != "" {
		lid, _ := strconv.Atoi(locationFilter)
		if lid > 0 {
			permitted := false
			for _, id := range permittedIDs {
				if id == lid {
					permitted = true
					break
				}
			}
			if !permitted {
				jsonErr(w, "Permission denied for this location", http.StatusForbidden)
				return
			}
			locationIDs = []int{lid}
		}
	}
	if len(locationIDs) == 0 {
		locationIDs = []int{currentLocationID}
	}

	// Base query builder — returns fresh *gorm.DB each time
	base := func() *gorm.DB {
		return h.db.Table("lab_ticket lt").
			Select(`lt.id_lab_ticket,
			lt.number_ticket,
			lt.tray,
			lt.invoice_id,
			i.date_create AS inv_date,
			lt.date_promise,
			lt.date_complete,
			COALESCE(EXTRACT(DAY FROM now() - lt.date_promise)::int, 0) AS late,
			lt.lab_ticket_status_id,
			lts.ticket_status,
			e.first_name AS emp_first,
			e.last_name AS emp_last,
			lt.patient_id,
			p.first_name AS pt_first,
			p.last_name AS pt_last,
			p.phone AS pt_phone,
			lt.notified,
			lt.dashboard_note,
			lt.g_or_c,
			l.id_location AS location_id,
			l.full_name AS location_name`).
			Joins("JOIN invoice i ON i.id_invoice = lt.invoice_id").
			Joins("JOIN location l ON l.id_location = i.location_id").
			Joins("LEFT JOIN lab_ticket_status lts ON lts.id_lab_ticket_status = lt.lab_ticket_status_id").
			Joins("LEFT JOIN employee e ON e.id_employee = lt.employee_id").
			Joins("LEFT JOIN patient p ON p.id_patient = lt.patient_id").
			Where("i.location_id IN ?", locationIDs).
			Where("lt.g_or_c = ?", gOrC)
	}

	// Apply filters
	applyFilters := func(tx *gorm.DB) *gorm.DB {
		if v := q.Get("status_id"); v != "" {
			if id, err := strconv.Atoi(v); err == nil {
				tx = tx.Where("lt.lab_ticket_status_id = ?", id)
			}
		}
			// g_or_c is set by endpoint, not query param
		if v := q.Get("search"); v != "" {
			s := "%" + strings.ToLower(v) + "%"
			tx = tx.Where("(LOWER(lt.number_ticket) LIKE ? OR LOWER(p.first_name||' '||p.last_name) LIKE ? OR LOWER(e.first_name||' '||e.last_name) LIKE ?)", s, s, s)
		}
		if v := q.Get("tray"); v != "" {
			tx = tx.Where("lt.tray = ?", v)
		}
		if v := q.Get("promised"); v != "" {
			today := time.Now().Format("2006-01-02")
			switch v {
			case "today":
				tx = tx.Where("lt.date_promise = ?", today)
			case "tomorrow":
				tm := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
				tx = tx.Where("lt.date_promise = ?", tm)
			case "overdue":
				tx = tx.Where("lt.date_promise < ?", today)
			}
		}
		if v := q.Get("complete"); v != "" {
			if v == "true" {
				tx = tx.Where("lt.date_complete IS NOT NULL")
			} else {
				tx = tx.Where("lt.date_complete IS NULL")
			}
		}
		return tx
	}

	// Sort
	sortCol := "lt.date_promise"
	sortDir := "ASC"
	if v := q.Get("sort"); v != "" {
		switch v {
		case "ticket":
			sortCol = "lt.number_ticket"
		case "late":
			sortCol = "late"
		case "date":
			sortCol = "inv_date"
		case "patient":
			sortCol = "pt_last"
		case "status":
			sortCol = "lts.ticket_status"
		case "employee":
			sortCol = "emp_last"
		}
	}
	if q.Get("order") == "desc" {
		sortDir = "DESC"
	}

	// CSV export (no pagination)
	if q.Get("output") == "csv" {
		var allRows []labStatusRow
		applyFilters(base()).Order(sortCol + " " + sortDir + " NULLS LAST").Find(&allRows)
		writeCSV(w, allRows)
		return
	}

	// Count total
	var total int64
	applyFilters(base()).Count(&total)

	// Pagination
	page, _ := strconv.Atoi(q.Get("page"))
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}

	var rows []labStatusRow
	applyFilters(base()).Order(sortCol + " " + sortDir + " NULLS LAST").
		Offset((page - 1) * perPage).Limit(perPage).Find(&rows)

	items := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		item := map[string]interface{}{
			"id_lab_ticket":  r.IDLabTicket,
			"number_ticket":  r.NumberTicket,
			"tray":           r.Tray,
			"invoice_id":     r.InvoiceID,
			"patient_id":     r.PatientID,
			"inv_date":       fmtDate(r.InvDate),
			"date_promise":   fmtDate(r.DatePromise),
			"date_complete":  fmtDate(r.DateComplete),
			"late":           r.Late,
			"status_id":      r.StatusID,
			"status":         r.StatusName,
			"employee":            strings.TrimSpace(r.EmployeeFirst + " " + r.EmployeeLast),
			"patient":        strings.TrimSpace(r.PatientLast + ", " + r.PatientFirst),
			"phone":          r.PatientPhone,
			"notified":       r.Notified,
			"dashboard_note": r.DashboardNote,
			"g_or_c":         r.GOrC,
			"location_id":    r.LocationID,
			"location_name":  r.LocationName,
		}
		items = append(items, item)
	}

	jsonResp(w, map[string]interface{}{
		"items":      items,
		"total":      total,
		"page":       page,
		"per_page":   perPage,
		"total_pages": int(math.Ceil(float64(total) / float64(perPage))),
	}, http.StatusOK)
}

// GET /api/orders/invoice-status
func (h *Handler) InvoiceStatus(w http.ResponseWriter, r *http.Request) {
	permittedIDs := middleware.PermittedLocationIDsFromCtx(r.Context())
	emp := middleware.EmployeeFromCtx(r.Context())

	var currentLocationID int
	if emp != nil && emp.LocationID != nil {
		currentLocationID = int(*emp.LocationID)
	}
	if currentLocationID == 0 && len(permittedIDs) > 0 {
		currentLocationID = permittedIDs[0]
	}

	q := r.URL.Query()

	locationFilter := q.Get("location_id")
	var locationIDs []int
	if locationFilter == "all" {
		locationIDs = permittedIDs
	} else if locationFilter != "" {
		lid, _ := strconv.Atoi(locationFilter)
		if lid > 0 {
			permitted := false
			for _, id := range permittedIDs {
				if id == lid {
					permitted = true
					break
				}
			}
			if !permitted {
				jsonErr(w, "Permission denied for this location", http.StatusForbidden)
				return
			}
			locationIDs = []int{lid}
		}
	}
	if len(locationIDs) == 0 {
		locationIDs = []int{currentLocationID}
	}

	type invRow struct {
		IDInvoice     int64      `gorm:"column:id_invoice"`
		NumberInvoice string     `gorm:"column:number_invoice"`
		DateCreate    *time.Time `gorm:"column:date_create"`
		Late          int        `gorm:"column:late"`
		StatusID      *int       `gorm:"column:status_invoice_id"`
		StatusName    *string    `gorm:"column:status_invoice_value"`
		EmployeeFirst string     `gorm:"column:emp_first"`
		EmployeeLast  string     `gorm:"column:emp_last"`
		PatientFirst  string     `gorm:"column:pt_first"`
		PatientLast   string     `gorm:"column:pt_last"`
		PatientPhone  *string    `gorm:"column:pt_phone"`
		Notified      *string    `gorm:"column:notified"`
		DashboardNote *string    `gorm:"column:dashboard_note"`
		LocationID    int        `gorm:"column:location_id"`
		LocationName  string     `gorm:"column:location_name"`
	}

	base := func() *gorm.DB {
		return h.db.Table("invoice i").
			Select(`i.id_invoice,
			i.number_invoice,
			i.date_create,
			COALESCE(EXTRACT(DAY FROM now() - i.date_create)::int, 0) AS late,
			i.status_invoice_id,
			si.status_invoice_value,
			e.first_name AS emp_first,
			e.last_name AS emp_last,
			lt.patient_id,
			p.first_name AS pt_first,
			p.last_name AS pt_last,
			p.phone AS pt_phone,
			i.notified,
			i.dashboard_note,
			l.id_location AS location_id,
			l.full_name AS location_name`).
			Joins("JOIN location l ON l.id_location = i.location_id").
			Joins("LEFT JOIN status_invoice si ON si.id_status_invoice = i.status_invoice_id").
			Joins("LEFT JOIN employee e ON e.id_employee = i.employee_id").
			Joins("LEFT JOIN patient p ON p.id_patient = i.patient_id").
			Where("i.location_id IN ?", locationIDs).
			Where("(si.status_type = 'patient' OR i.status_invoice_id IS NULL)")
	}

	applyFilters := func(tx *gorm.DB) *gorm.DB {
		if v := q.Get("status_id"); v != "" {
			if id, err := strconv.Atoi(v); err == nil {
				tx = tx.Where("i.status_invoice_id = ?", id)
			}
		}
		if v := q.Get("search"); v != "" {
			s := "%" + strings.ToLower(v) + "%"
			tx = tx.Where("(LOWER(i.number_invoice) LIKE ? OR LOWER(p.first_name||' '||p.last_name) LIKE ? OR LOWER(e.first_name||' '||e.last_name) LIKE ?)", s, s, s)
		}
		// No date_promise/date_complete on invoice
		return tx
	}

	sortCol := "i.date_create"
	sortDir := "DESC"
	if v := q.Get("sort"); v != "" {
		switch v {
		case "invoice":
			sortCol = "i.number_invoice"
		case "late":
			sortCol = "late"
		case "date":
			sortCol = "i.date_create"
		case "patient":
			sortCol = "pt_last"
		case "status":
			sortCol = "si.status_invoice_value"
		case "employee":
			sortCol = "emp_last"
		}
	}
	if q.Get("order") == "asc" {
		sortDir = "ASC"
	}

	if q.Get("output") == "csv" {
		var allRows []invRow
		applyFilters(base()).Order(sortCol + " " + sortDir + " NULLS LAST").Find(&allRows)

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=invoice-status.csv")
		cw := csv.NewWriter(w)
		cw.Write([]string{"invoice", "inv_date", "rep", "late", "status", "patient", "phone", "notified", "dashboard_note"})
		for _, r := range allRows {
			cw.Write([]string{
				r.NumberInvoice, fmtDateStr(r.DateCreate),
				strings.TrimSpace(r.EmployeeFirst + " " + r.EmployeeLast),
				strconv.Itoa(r.Late),
				ptrStr(r.StatusName),
				strings.TrimSpace(r.PatientLast + ", " + r.PatientFirst),
				ptrStr(r.PatientPhone), ptrStr(r.Notified), ptrStr(r.DashboardNote),
			})
		}
		cw.Flush()
		return
	}

	var total int64
	applyFilters(base()).Count(&total)

	page, _ := strconv.Atoi(q.Get("page"))
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}

	var rows []invRow
	applyFilters(base()).Order(sortCol + " " + sortDir + " NULLS LAST").
		Offset((page - 1) * perPage).Limit(perPage).Find(&rows)

	items := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		items = append(items, map[string]interface{}{
			"id_invoice":     r.IDInvoice,
			"number_invoice": r.NumberInvoice,
			"inv_date":       fmtDate(r.DateCreate),
			"late":           r.Late,
			"status_id":      r.StatusID,
			"status":         r.StatusName,
			"employee":            strings.TrimSpace(r.EmployeeFirst + " " + r.EmployeeLast),
			"patient":        strings.TrimSpace(r.PatientLast + ", " + r.PatientFirst),
			"phone":          r.PatientPhone,
			"notified":       r.Notified,
			"dashboard_note": r.DashboardNote,
			"location_id":    r.LocationID,
			"location_name":  r.LocationName,
		})
	}

	jsonResp(w, map[string]interface{}{
		"items":       items,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": int(math.Ceil(float64(total) / float64(perPage))),
	}, http.StatusOK)
}

// POST /api/orders/ticket/{id}/dashboard-note
func (h *Handler) UpdateTicketNote(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if id == 0 {
		jsonErr(w, "invalid ticket id", http.StatusBadRequest)
		return
	}
	var body struct {
		Note *string `json:"dashboard_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	res := h.db.Table("lab_ticket").Where("id_lab_ticket = ?", id).Update("dashboard_note", body.Note)
	if res.Error != nil {
		jsonErr(w, res.Error.Error(), http.StatusInternalServerError)
		return
	}
	if res.RowsAffected == 0 {
		jsonErr(w, "ticket not found", http.StatusNotFound)
		return
	}
	jsonResp(w, map[string]interface{}{"message": "note updated"}, http.StatusOK)
}

// POST /api/orders/invoice/{id}/dashboard-note
func (h *Handler) UpdateInvoiceNote(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if id == 0 {
		jsonErr(w, "invalid invoice id", http.StatusBadRequest)
		return
	}
	var body struct {
		Note *string `json:"dashboard_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	res := h.db.Table("invoice").Where("id_invoice = ?", id).Update("dashboard_note", body.Note)
	if res.Error != nil {
		jsonErr(w, res.Error.Error(), http.StatusInternalServerError)
		return
	}
	if res.RowsAffected == 0 {
		jsonErr(w, "invoice not found", http.StatusNotFound)
		return
	}
	jsonResp(w, map[string]interface{}{"message": "note updated"}, http.StatusOK)
}

// GET /api/orders/locations
func (h *Handler) Locations(w http.ResponseWriter, r *http.Request) {
	permittedIDs := middleware.PermittedLocationIDsFromCtx(r.Context())
	emp := middleware.EmployeeFromCtx(r.Context())

	var currentLocationID int
	if emp != nil && emp.LocationID != nil {
		currentLocationID = int(*emp.LocationID)
	}

	type locRow struct {
		IDLocation int    `gorm:"column:id_location"`
		FullName   string `gorm:"column:full_name"`
	}
	var locs []locRow
	h.db.Table("location").
		Select("id_location, full_name").
		Where("id_location IN ? AND store_active = true", permittedIDs).
		Order("full_name ASC").
		Find(&locs)

	items := make([]map[string]interface{}, 0, len(locs))
	for _, l := range locs {
		items = append(items, map[string]interface{}{
			"location_id": l.IDLocation,
			"full_name":   l.FullName,
			"current":     l.IDLocation == currentLocationID,
		})
	}
	jsonResp(w, items, http.StatusOK)
}

// GET /api/orders/sms-templates
func (h *Handler) SMSTemplates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	tx := h.db.Table("sms_template").Where("active = true")
	if category != "" {
		tx = tx.Where("category = ?", category)
	}
	type row struct {
		ID       int    `gorm:"column:id_sms_template"`
		Category string `gorm:"column:category"`
		Name     string `gorm:"column:name"`
		Body     string `gorm:"column:body"`
		IsSystem bool   `gorm:"column:is_system"`
	}
	var rows []row
	tx.Order("category, name").Find(&rows)
	items := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		items = append(items, map[string]interface{}{
			"id":        r.ID,
			"category":  r.Category,
			"name":      r.Name,
			"body":      r.Body,
			"is_system": r.IsSystem,
		})
	}
	jsonResp(w, items, http.StatusOK)
}

// POST /api/orders/send-sms
func (h *Handler) SendSMS(w http.ResponseWriter, r *http.Request) {
	emp := middleware.EmployeeFromCtx(r.Context())

	var body struct {
		Phone      string            `json:"phone"`
		PatientID  *int64            `json:"patient_id"`
		TemplateID *int              `json:"template_id"`
		Message    *string           `json:"message"`
		Vars       map[string]string `json:"vars"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Phone == "" {
		jsonErr(w, "phone is required", http.StatusBadRequest)
		return
	}

	var smsMessage string
	if body.TemplateID != nil {
		var tpl struct {
			Body string `gorm:"column:body"`
		}
		if err := h.db.Table("sms_template").Where("id_sms_template = ? AND active = true", *body.TemplateID).Scan(&tpl).Error; err != nil || tpl.Body == "" {
			jsonErr(w, "template not found", http.StatusNotFound)
			return
		}
		rendered, err := communication.RenderSMSTemplate(tpl.Body, body.Vars)
		if err != nil {
			jsonErr(w, "template render error: "+err.Error(), http.StatusBadRequest)
			return
		}
		smsMessage = rendered
	} else if body.Message != nil && *body.Message != "" {
		smsMessage = *body.Message
	} else {
		jsonErr(w, "template_id or message is required", http.StatusBadRequest)
		return
	}

	res := communication.SendSMS(body.Phone, smsMessage)
	if res.Status != "accepted" {
		jsonErr(w, "SMS failed: "+res.Error, http.StatusBadGateway)
		return
	}

	// Log communication
	if body.PatientID != nil && emp != nil {
		var locID int
		if emp.LocationID != nil {
			locID = int(*emp.LocationID)
		}
		h.logSMS(*body.PatientID, smsMessage, "SMS sent from orders dashboard", int(emp.IDEmployee), locID)
	}

	jsonResp(w, map[string]interface{}{"message": "SMS sent", "status": "accepted"}, http.StatusOK)
}

func (h *Handler) logSMS(patientID int64, content, description string, empID, locID int) {
	var commTypeID int
	h.db.Table("communication_type").Where("communication_type = ?", "SMS").Pluck("id_communication_type", &commTypeID)
	if commTypeID == 0 {
		return
	}
	h.db.Exec(`INSERT INTO patient_communication_history
		(patient_id, communication_type_id, content, description, employee_id, location_id, date_communication)
		VALUES (?, ?, ?, ?, ?, ?, now())`,
		patientID, commTypeID, content, description, empID, locID)
}

// ── helpers ──

func fmtDate(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02")
}

func writeCSV(w http.ResponseWriter, rows []labStatusRow) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=lab-status.csv")

	cw := csv.NewWriter(w)
	cw.Write([]string{"lab_ticket", "tray", "inv_date", "rep", "late", "promised", "complete", "status", "patient", "phone", "notified", "dashboard_note"})

	for _, r := range rows {
		cw.Write([]string{
			r.NumberTicket,
			ptrStr(r.Tray),
			fmtDateStr(r.InvDate),
			strings.TrimSpace(r.EmployeeFirst + " " + r.EmployeeLast),
			strconv.Itoa(r.Late),
			fmtDateStr(r.DatePromise),
			fmtDateStr(r.DateComplete),
			r.StatusName,
			strings.TrimSpace(r.PatientLast + ", " + r.PatientFirst),
			ptrStr(r.PatientPhone),
			ptrStr(r.Notified),
			ptrStr(r.DashboardNote),
		})
	}
	cw.Flush()
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func fmtDateStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("01/02/2006")
}

func jsonResp(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}
