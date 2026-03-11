package task_service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	invModel "sighthub-backend/internal/models/inventory"
	invoiceModel "sighthub-backend/internal/models/invoices"
	taskModel "sighthub-backend/internal/models/tasks"
	"sighthub-backend/pkg/sku"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── DTOs ──────────────────────────────────────────────────────────────────────

type EmployeeItem struct {
	EmployeeID int    `json:"employee_id"`
	Name       string `json:"name"`
}

type TaskListItem struct {
	IDTask         int64   `json:"id_task"`
	Title          string  `json:"title"`
	Description    *string `json:"description"`
	Status         string  `json:"status"`
	Priority       string  `json:"priority"`
	DueDate        *string `json:"due_date"`
	CreateAt       *string `json:"create_at"`
	FromEmployee   *string `json:"from_employee"`
	FromEmployeeID *int    `json:"from_employee_id"`
	ForEmployee    *string `json:"for_employee"`
	ForEmployeeID  *int    `json:"for_employee_id"`
}

type ResourceItem struct {
	IDTasksResource   int64   `json:"id_tasks_resource"`
	TaskID            int64   `json:"task_id"`
	InvoiceID         *int64  `json:"invoice_id"`
	InvoiceItemSaleID *int64  `json:"invoice_item_sale_id"`
	PatientID         *int64  `json:"patient_id"`
	SKU               *int64  `json:"sku"`
	AddPath           *string `json:"add_path"`
}

type HistoryItem struct {
	IDTaskHistory     int64   `json:"id_task_history"`
	Description       *string `json:"description"`
	LastUpdate        *string `json:"last_update"`
	OldEmployeeID     *int    `json:"old_employee_id"`
	NewEmployeeID     *int    `json:"new_employee_id"`
	OldStatus         *string `json:"old_status"`
	NewStatus         *string `json:"new_status"`
	OldTaskResourceID *int64  `json:"old_task_resource_id"`
	NewTaskResourceID *int64  `json:"new_task_resource_id"`
}

type CommentItem struct {
	IDTaskComment int64   `json:"id_task_comment"`
	AuthorID      int     `json:"author_id"`
	AuthorName    *string `json:"author_name"`
	Message       string  `json:"message"`
	CreatedAt     *string `json:"created_at"`
}

type TaskDetail struct {
	Task      TaskListItem   `json:"task"`
	Resources []ResourceItem `json:"resources"`
	History   []HistoryItem  `json:"history"`
	Comments  []CommentItem  `json:"comments"`
}

type ResourceInput struct {
	InvoiceID         interface{} `json:"invoice_id"`
	InvoiceItemSaleID *int64      `json:"invoice_item_sale_id"`
	PatientID         *int64      `json:"patient_id"`
	InventoryID       *int64      `json:"inventory_id"`
	SKU               *string     `json:"sku"`
	AddPath           *string     `json:"add_path"`
}

type CreateTaskInput struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	EmployeeID  *int            `json:"employee_id"`
	DueDate     *string         `json:"due_date"`
	Priority    string          `json:"priority"`
	Resources   []ResourceInput `json:"resources"`
	CreatorID   int             // set by handler from JWT
}

type UpdateTaskInput struct {
	EmployeeID  *int            `json:"employee_id"`
	Status      *string         `json:"status"`
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	DueDate     *string         `json:"due_date"`
	Priority    *string         `json:"priority"`
	Resources   []ResourceInput `json:"resources"`
}

type InvoiceResult struct {
	IDInvoice     int64  `json:"id_invoice"`
	NumberInvoice string `json:"number_invoice"`
	RedirectURL   string `json:"redirect_url"`
}

type CommentResult struct {
	CommentID    int64   `json:"comment_id"`
	EmployeeName string  `json:"employee_name"`
	Datetime     *string `json:"datetime"`
	Content      string  `json:"content"`
}

// ─── Employee helpers ──────────────────────────────────────────────────────────

// GetEmployeeByUsername looks up the Employee + login by JWT username.
func (s *Service) GetEmployeeByUsername(username string) (*empModel.Employee, *authModel.EmployeeLogin, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &login, nil
		}
		return nil, &login, err
	}
	return &emp, &login, nil
}

// GetEmployees returns a simple list of all employees.
func (s *Service) GetEmployees() ([]EmployeeItem, error) {
	var emps []empModel.Employee
	if err := s.db.Select("id_employee, first_name, last_name").Find(&emps).Error; err != nil {
		return nil, err
	}
	out := make([]EmployeeItem, len(emps))
	for i, e := range emps {
		out[i] = EmployeeItem{
			EmployeeID: e.IDEmployee,
			Name:       strings.TrimSpace(e.FirstName + " " + e.LastName),
		}
	}
	return out, nil
}

// ─── Task list ─────────────────────────────────────────────────────────────────

type GetTasksFilters struct {
	Mode       string   // "for_me" | "by_me" | "all"
	Statuses   []string
	DateStart  *string
	DateEnd    *string
	EmployeeID int
}

func (s *Service) GetTasks(f GetTasksFilters) ([]TaskListItem, error) {
	var dateStart, dateEnd *time.Time

	// Auto-filter: single "done" status without dates → last 30 days
	if len(f.Statuses) == 1 && f.Statuses[0] == "done" && f.DateStart == nil && f.DateEnd == nil {
		end := time.Now().UTC()
		start := end.AddDate(0, 0, -30)
		dateEnd = &end
		dateStart = &start
	} else {
		if f.DateStart != nil {
			t, err := time.Parse("2006-01-02", *f.DateStart)
			if err != nil {
				return nil, fmt.Errorf("invalid date_start format: use YYYY-MM-DD")
			}
			dateStart = &t
		}
		if f.DateEnd != nil {
			t, err := time.Parse("2006-01-02", *f.DateEnd)
			if err != nil {
				return nil, fmt.Errorf("invalid date_end format: use YYYY-MM-DD")
			}
			endOfDay := t.Add(24 * time.Hour)
			dateEnd = &endOfDay
		}
	}

	query := s.db.Model(&taskModel.Task{})

	switch f.Mode {
	case "by_me":
		query = query.Where("creator_id = ?", f.EmployeeID)
	case "all":
		// no employee filter
	default: // "for_me"
		query = query.Where("employee_id = ?", f.EmployeeID)
	}

	if len(f.Statuses) > 0 {
		query = query.Where("status IN ?", f.Statuses)
	}
	if dateStart != nil {
		query = query.Where("create_at >= ?", *dateStart)
	}
	if dateEnd != nil {
		query = query.Where("create_at <= ?", *dateEnd)
	}

	var tasks []taskModel.Task
	if err := query.Order("create_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	empIDs := map[int]struct{}{}
	for _, t := range tasks {
		empIDs[t.CreatorID] = struct{}{}
		if t.EmployeeID != nil {
			empIDs[*t.EmployeeID] = struct{}{}
		}
	}
	empMap := s.loadEmployeeNames(empIDs)

	out := make([]TaskListItem, len(tasks))
	for i, t := range tasks {
		out[i] = buildTaskListItem(t, empMap)
	}
	return out, nil
}

// ─── Task detail ───────────────────────────────────────────────────────────────

func (s *Service) GetTask(taskID int64) (*TaskDetail, error) {
	var task taskModel.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	empIDs := map[int]struct{}{task.CreatorID: {}}
	if task.EmployeeID != nil {
		empIDs[*task.EmployeeID] = struct{}{}
	}
	empMap := s.loadEmployeeNames(empIDs)

	var resources []taskModel.TasksResource
	s.db.Where("task_id = ?", taskID).Find(&resources)
	resItems := make([]ResourceItem, len(resources))
	for i, r := range resources {
		resItems[i] = ResourceItem{
			IDTasksResource:   r.IDTasksResource,
			TaskID:            r.TaskID,
			InvoiceID:         r.InvoiceID,
			InvoiceItemSaleID: r.InvoiceItemSaleID,
			PatientID:         r.PatientID,
			SKU:               r.InventoryID,
			AddPath:           r.AddPath,
		}
	}

	var histories []taskModel.TaskHistory
	s.db.Where("task_id = ?", taskID).Order("last_update DESC").Find(&histories)
	histItems := make([]HistoryItem, len(histories))
	for i, h := range histories {
		var lu *string
		if h.LastUpdate != nil {
			str := h.LastUpdate.Format(time.RFC3339)
			lu = &str
		}
		histItems[i] = HistoryItem{
			IDTaskHistory:     h.IDTaskHistory,
			Description:       h.Description,
			LastUpdate:        lu,
			OldEmployeeID:     h.OldEmployeeID,
			NewEmployeeID:     h.NewEmployeeID,
			OldStatus:         h.OldStatus,
			NewStatus:         h.NewStatus,
			OldTaskResourceID: h.OldTaskResourceID,
			NewTaskResourceID: h.NewTaskResourceID,
		}
	}

	var comments []taskModel.TaskComment
	s.db.Where("task_id = ?", taskID).Order("created_at ASC").Find(&comments)
	commentItems := make([]CommentItem, len(comments))
	for i, c := range comments {
		var ca *string
		if c.CreatedAt != nil {
			str := c.CreatedAt.Format(time.RFC3339)
			ca = &str
		}
		authorName := s.getEmployeeName(c.AuthorID)
		commentItems[i] = CommentItem{
			IDTaskComment: c.IDTaskComment,
			AuthorID:      c.AuthorID,
			AuthorName:    authorName,
			Message:       c.Message,
			CreatedAt:     ca,
		}
	}

	return &TaskDetail{
		Task:      buildTaskListItem(task, empMap),
		Resources: resItems,
		History:   histItems,
		Comments:  commentItems,
	}, nil
}

// ─── Create task ───────────────────────────────────────────────────────────────

func (s *Service) CreateTask(input CreateTaskInput) (*taskModel.Task, error) {
	priority := input.Priority
	if priority == "" {
		priority = "medium"
	}
	employeeID := input.EmployeeID
	if employeeID == nil {
		employeeID = &input.CreatorID
	}

	var dueDate *time.Time
	if input.DueDate != nil && *input.DueDate != "" {
		t, err := time.Parse(time.RFC3339, *input.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date format. Use ISO 8601 format like 2025-08-01T17:00:00")
		}
		dueDate = &t
	}

	description := input.Description

	parsed, invoiceNumbers, err := s.parseResources(input.Resources)
	if err != nil {
		return nil, err
	}
	for _, num := range invoiceNumbers {
		tag := fmt.Sprintf("Invoice #%s", num)
		if !strings.Contains(description, tag) {
			if description != "" {
				description += "\n"
			}
			description += tag
		}
	}

	task := &taskModel.Task{
		Title:      input.Title,
		Status:     "to do",
		Priority:   priority,
		CreatorID:  input.CreatorID,
		EmployeeID: employeeID,
		DueDate:    dueDate,
	}
	if description != "" {
		task.Description = &description
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		for i := range parsed {
			parsed[i].TaskID = task.IDTask
			if err := tx.Create(&parsed[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}

// ─── Update task ───────────────────────────────────────────────────────────────

func (s *Service) UpdateTask(taskID int64, input UpdateTaskInput) (*taskModel.Task, error) {
	var task taskModel.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	type change struct{ old, new interface{} }
	changes := map[string]change{}

	if input.EmployeeID != nil && (task.EmployeeID == nil || *input.EmployeeID != *task.EmployeeID) {
		changes["employee"] = change{task.EmployeeID, input.EmployeeID}
		task.EmployeeID = input.EmployeeID
	}
	if input.Status != nil && *input.Status != task.Status {
		changes["status"] = change{task.Status, *input.Status}
		task.Status = *input.Status
	}
	if input.Title != nil && *input.Title != task.Title {
		changes["title"] = change{task.Title, *input.Title}
		task.Title = *input.Title
	}
	if input.Description != nil {
		oldDesc := ""
		if task.Description != nil {
			oldDesc = *task.Description
		}
		if *input.Description != oldDesc {
			changes["description"] = change{task.Description, *input.Description}
			task.Description = input.Description
		}
	}
	if input.DueDate != nil {
		var newDue *time.Time
		if *input.DueDate != "" {
			t, err := time.Parse(time.RFC3339, *input.DueDate)
			if err != nil {
				return nil, fmt.Errorf("invalid due_date format. Expected ISO format")
			}
			newDue = &t
		}
		oldStr, newStr := "", ""
		if task.DueDate != nil {
			oldStr = task.DueDate.Format(time.RFC3339)
		}
		if newDue != nil {
			newStr = newDue.Format(time.RFC3339)
		}
		if oldStr != newStr {
			changes["due_date"] = change{oldStr, newStr}
			task.DueDate = newDue
		}
	}
	if input.Priority != nil && *input.Priority != task.Priority {
		changes["priority"] = change{task.Priority, *input.Priority}
		task.Priority = *input.Priority
	}

	parsed, invoiceNumbers, err := s.parseResources(input.Resources)
	if err != nil {
		return nil, err
	}
	for _, num := range invoiceNumbers {
		tag := fmt.Sprintf("Invoice #%s", num)
		desc := ""
		if task.Description != nil {
			desc = *task.Description
		}
		if !strings.Contains(desc, tag) {
			if desc != "" {
				desc += "\n"
			}
			desc += tag
			task.Description = &desc
		}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&task).Error; err != nil {
			return err
		}
		for i := range parsed {
			parsed[i].TaskID = taskID
			if err := tx.Create(&parsed[i]).Error; err != nil {
				return err
			}
			rel := taskModel.TaskResourceRelation{
				TaskID:          taskID,
				TasksResourceID: parsed[i].IDTasksResource,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		if len(changes) > 0 {
			var oldEmpID, newEmpID *int
			var oldStatus, newStatus *string
			if v, ok := changes["employee"]; ok {
				if v.old != nil {
					id := v.old.(*int)
					oldEmpID = id
				}
				if v.new != nil {
					id := v.new.(*int)
					newEmpID = id
				}
			}
			if v, ok := changes["status"]; ok {
				s1 := v.old.(string)
				s2 := v.new.(string)
				oldStatus = &s1
				newStatus = &s2
			}
			desc := "Task updated"
			history := &taskModel.TaskHistory{
				TaskID:        taskID,
				OldEmployeeID: oldEmpID,
				NewEmployeeID: newEmpID,
				OldStatus:     oldStatus,
				NewStatus:     newStatus,
				Description:   &desc,
			}
			if err := tx.Create(history).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ─── Delete task ───────────────────────────────────────────────────────────────

func (s *Service) DeleteTask(taskID int64) error {
	var task taskModel.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return err
	}
	if strings.TrimSpace(strings.ToLower(task.Status)) != "to do" {
		return fmt.Errorf("only tasks with status TO DO can be deleted")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var histCount int64
		tx.Model(&taskModel.TaskHistory{}).Where("task_id = ?", taskID).Count(&histCount)
		if histCount > 0 {
			oldStatus := task.Status
			newStatus := "deleted"
			desc := "Task deleted"
			h := &taskModel.TaskHistory{
				TaskID:        taskID,
				OldEmployeeID: task.EmployeeID,
				NewEmployeeID: nil,
				Description:   &desc,
				OldStatus:     &oldStatus,
				NewStatus:     &newStatus,
			}
			if err := tx.Create(h).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&taskModel.Task{}, taskID).Error
	})
}

// ─── Delete task resource ──────────────────────────────────────────────────────

func (s *Service) DeleteTaskResource(resourceID int64) error {
	var res taskModel.TasksResource
	if err := s.db.First(&res, resourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return err
	}
	return s.db.Delete(&taskModel.TasksResource{}, resourceID).Error
}

// ─── Invoice search ────────────────────────────────────────────────────────────

func (s *Service) SearchInvoice(query string, locationID int64) ([]InvoiceResult, error) {
	if query == "" {
		return nil, fmt.Errorf("invoice number not provided")
	}

	param := "%" + query + "%"
	var invoices []invoiceModel.Invoice
	if isDigits(query) {
		s.db.Where("CAST(id_invoice AS VARCHAR) ILIKE ? OR number_invoice ILIKE ?", param, param).
			Find(&invoices)
	} else {
		s.db.Where("number_invoice ILIKE ?", param).Find(&invoices)
	}

	if len(invoices) == 0 {
		return nil, fmt.Errorf("no invoices found")
	}

	var out []InvoiceResult
	for _, inv := range invoices {
		num := inv.NumberInvoice
		if strings.HasPrefix(num, "V") {
			if locationID == inv.LocationID {
				out = append(out, InvoiceResult{
					IDInvoice:     inv.IDInvoice,
					NumberInvoice: num,
					RedirectURL:   fmt.Sprintf("/inventory/receipts/vendors/invoice/%d/%d/*", inv.IDInvoice, inv.VendorID),
				})
			}
		} else if strings.HasPrefix(num, "I") {
			if locationID == inv.LocationID {
				out = append(out, InvoiceResult{
					IDInvoice:     inv.IDInvoice,
					NumberInvoice: num,
					RedirectURL:   fmt.Sprintf("/inventory/transfers/invoice/%d", inv.IDInvoice),
				})
			} else if inv.ToLocationID != nil && locationID == *inv.ToLocationID {
				out = append(out, InvoiceResult{
					IDInvoice:     inv.IDInvoice,
					NumberInvoice: num,
					RedirectURL:   fmt.Sprintf("/receipts/store/invoice/%d", inv.IDInvoice),
				})
			}
		} else if strings.HasPrefix(num, "S") {
			if locationID == inv.LocationID {
				out = append(out, InvoiceResult{
					IDInvoice:     inv.IDInvoice,
					NumberInvoice: num,
					RedirectURL:   fmt.Sprintf("/patient/%d/invoice/%d", inv.PatientID, inv.IDInvoice),
				})
			}
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no accessible invoices found")
	}
	return out, nil
}

// ─── Comments ──────────────────────────────────────────────────────────────────

func (s *Service) CreateComment(taskID int64, authorID int, message string) (*CommentResult, error) {
	var task taskModel.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	comment := &taskModel.TaskComment{
		TaskID:   taskID,
		AuthorID: authorID,
		Message:  message,
	}
	if err := s.db.Create(comment).Error; err != nil {
		return nil, err
	}

	var emp empModel.Employee
	s.db.First(&emp, authorID)
	name := strings.TrimSpace(emp.FirstName + " " + emp.LastName)

	var dt *string
	if comment.CreatedAt != nil {
		str := comment.CreatedAt.Format(time.RFC3339)
		dt = &str
	}
	return &CommentResult{
		CommentID:    comment.IDTaskComment,
		EmployeeName: name,
		Datetime:     dt,
		Content:      comment.Message,
	}, nil
}

func (s *Service) UpdateComment(commentID int64, authorID int, message string) error {
	var c taskModel.TaskComment
	if err := s.db.First(&c, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return err
	}
	if c.AuthorID != authorID {
		return fmt.Errorf("forbidden")
	}
	return s.db.Model(&c).Update("message", message).Error
}

func (s *Service) DeleteComment(commentID int64, authorID int) error {
	var c taskModel.TaskComment
	if err := s.db.First(&c, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return err
	}
	if c.AuthorID != authorID {
		return fmt.Errorf("forbidden")
	}
	return s.db.Delete(&taskModel.TaskComment{}, commentID).Error
}

// ─── Private helpers ──────────────────────────────────────────────────────────

var invoicePathRe = regexp.MustCompile(`/invoice/(\d+)`)

func extractInvoiceIDFromPath(path string) *int64 {
	if path == "" {
		return nil
	}
	m := invoicePathRe.FindStringSubmatch(path)
	if m == nil {
		return nil
	}
	var id int64
	fmt.Sscanf(m[1], "%d", &id)
	return &id
}

func (s *Service) resolveInvoiceIdentifier(identifier interface{}) *int64 {
	if identifier == nil {
		return nil
	}
	switch v := identifier.(type) {
	case float64: // JSON numbers come as float64
		id := int64(v)
		return &id
	case int:
		id := int64(v)
		return &id
	case int64:
		return &v
	case string:
		if isDigits(v) {
			var id int64
			fmt.Sscanf(v, "%d", &id)
			return &id
		}
		var inv invoiceModel.Invoice
		if err := s.db.Where("number_invoice = ?", v).First(&inv).Error; err == nil {
			return &inv.IDInvoice
		}
		return nil
	}
	return nil
}

func (s *Service) parseResources(resources []ResourceInput) ([]taskModel.TasksResource, []string, error) {
	var parsed []taskModel.TasksResource
	var invoiceNumbers []string

	for _, res := range resources {
		var invoiceID *int64
		if res.InvoiceID != nil {
			invoiceID = s.resolveInvoiceIdentifier(res.InvoiceID)
		} else if res.AddPath != nil {
			invoiceID = extractInvoiceIDFromPath(*res.AddPath)
		}

		if invoiceID != nil {
			var inv invoiceModel.Invoice
			if err := s.db.First(&inv, invoiceID).Error; err == nil {
				invoiceNumbers = append(invoiceNumbers, inv.NumberInvoice)
			}
		}

		inventoryID := res.InventoryID
		if inventoryID == nil && res.SKU != nil {
			normalized := sku.Normalize(*res.SKU)
			var item invModel.Inventory
			if err := s.db.Where("sku = ?", normalized).First(&item).Error; err == nil {
				inventoryID = &item.IDInventory
			}
		}

		parsed = append(parsed, taskModel.TasksResource{
			InvoiceID:         invoiceID,
			InvoiceItemSaleID: res.InvoiceItemSaleID,
			PatientID:         res.PatientID,
			InventoryID:       inventoryID,
			AddPath:           res.AddPath,
		})
	}
	return parsed, invoiceNumbers, nil
}

func (s *Service) loadEmployeeNames(ids map[int]struct{}) map[int]empModel.Employee {
	if len(ids) == 0 {
		return nil
	}
	idSlice := make([]int, 0, len(ids))
	for id := range ids {
		idSlice = append(idSlice, id)
	}
	var emps []empModel.Employee
	s.db.Where("id_employee IN ?", idSlice).Find(&emps)
	m := make(map[int]empModel.Employee, len(emps))
	for _, e := range emps {
		m[e.IDEmployee] = e
	}
	return m
}

func (s *Service) getEmployeeName(id int) *string {
	var emp empModel.Employee
	if err := s.db.First(&emp, id).Error; err != nil {
		return nil
	}
	name := strings.TrimSpace(emp.FirstName + " " + emp.LastName)
	return &name
}

func buildTaskListItem(t taskModel.Task, empMap map[int]empModel.Employee) TaskListItem {
	item := TaskListItem{
		IDTask:      t.IDTask,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
	}
	if t.DueDate != nil {
		str := t.DueDate.Format(time.RFC3339)
		item.DueDate = &str
	}
	if t.CreateAt != nil {
		str := t.CreateAt.Format(time.RFC3339)
		item.CreateAt = &str
	}
	if empMap != nil {
		if creator, ok := empMap[t.CreatorID]; ok {
			name := strings.TrimSpace(creator.FirstName + " " + creator.LastName)
			item.FromEmployee = &name
			item.FromEmployeeID = &creator.IDEmployee
		}
		if t.EmployeeID != nil {
			if assignee, ok := empMap[*t.EmployeeID]; ok {
				name := strings.TrimSpace(assignee.FirstName + " " + assignee.LastName)
				item.ForEmployee = &name
				item.ForEmployeeID = &assignee.IDEmployee
			}
		}
	}
	return item
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
