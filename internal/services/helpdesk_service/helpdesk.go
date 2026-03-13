package helpdesk_service

import (
	"bytes"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/config"
	genModel "sighthub-backend/internal/models/general"
	pkgActivity "sighthub-backend/pkg/activitylog"
)

type Service struct {
	db  *gorm.DB
	cfg *config.Config
}

func New(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

type ScreenshotItem struct {
	IDHelpdeskTicketScreenshot int64     `json:"id_helpdesk_ticket_screenshot"`
	URL                        string    `json:"url"`
	CreatedAt                  time.Time `json:"created_at"`
}

type TicketResult struct {
	IDHelpdeskTicket int64            `json:"id_helpdesk_ticket"`
	EmployeeLogin    string           `json:"employee_login"`
	Subject          string           `json:"subject"`
	Description      string           `json:"description"`
	SourceDomain     *string          `json:"source_domain"`
	Status           string           `json:"status"`
	CreatedAt        string           `json:"created_at"`
	ForwardedAt      *string          `json:"forwarded_at"`
	ForwardOk        bool             `json:"forward_ok"`
	ForwardError     *string          `json:"forward_error"`
	Screenshots      []ScreenshotItem `json:"screenshots"`
}

type ListResult struct {
	Tickets []TicketResult `json:"tickets"`
	Meta    PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
	Pages   int   `json:"pages"`
	HasNext bool  `json:"has_next"`
	HasPrev bool  `json:"has_prev"`
}

// ── Methods ───────────────────────────────────────────────────────────────────

// ListTickets returns paginated tickets for an employee login.
func (s *Service) ListTickets(employeeLogin string, page, perPage int, status, q, order string) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	base := s.db.Model(&genModel.HelpdeskTicket{}).Where("employee_login = ?", employeeLogin)

	if status != "" {
		base = base.Where("status = ?", status)
	}
	if sq := strings.TrimSpace(q); sq != "" {
		like := "%" + sq + "%"
		base = base.Where("subject ILIKE ? OR description ILIKE ?", like, like)
	}

	var total int64
	base.Count(&total)

	if strings.ToLower(order) == "asc" {
		base = base.Order("created_at ASC, id_helpdesk_ticket ASC")
	} else {
		base = base.Order("created_at DESC, id_helpdesk_ticket DESC")
	}

	offset := (page - 1) * perPage
	var tickets []genModel.HelpdeskTicket
	if err := base.Offset(offset).Limit(perPage).Find(&tickets).Error; err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(tickets))
	for _, t := range tickets {
		ids = append(ids, t.IDHelpdeskTicket)
	}
	screenshotMap := s.loadScreenshots(ids)

	items := make([]TicketResult, 0, len(tickets))
	for _, t := range tickets {
		items = append(items, toTicketResult(t, screenshotMap[t.IDHelpdeskTicket]))
	}

	pages := int((total + int64(perPage) - 1) / int64(perPage))
	return &ListResult{
		Tickets: items,
		Meta: PaginationMeta{
			Page:    page,
			PerPage: perPage,
			Total:   total,
			Pages:   pages,
			HasNext: page < pages,
			HasPrev: page > 1,
		},
	}, nil
}

// GetTicket returns a single ticket belonging to the given employee login.
func (s *Service) GetTicket(id int64, employeeLogin string) (*TicketResult, error) {
	var ticket genModel.HelpdeskTicket
	if err := s.db.Where("id_helpdesk_ticket = ? AND employee_login = ?", id, employeeLogin).
		First(&ticket).Error; err != nil {
		return nil, fmt.Errorf("ticket not found")
	}
	shots := s.loadScreenshots([]int64{id})
	r := toTicketResult(ticket, shots[id])
	return &r, nil
}

// CreateTicket saves a new ticket, logs activity, and forwards to external helpdesk.
func (s *Service) CreateTicket(employeeLogin, subject, description string, screenshots []string, sourceDomain *string) (*TicketResult, error) {
	ticket := genModel.HelpdeskTicket{
		EmployeeLogin: employeeLogin,
		Subject:       subject,
		Description:   description,
		SourceDomain:  sourceDomain,
	}
	if err := s.db.Create(&ticket).Error; err != nil {
		return nil, err
	}

	for _, u := range screenshots {
		s.db.Create(&genModel.HelpdeskTicketScreenshot{TicketID: ticket.IDHelpdeskTicket, URL: u})
	}

	pkgActivity.Log(s.db, "helpdesk", "ticket_create",
		pkgActivity.WithEntity(ticket.IDHelpdeskTicket),
		pkgActivity.WithDetails(map[string]interface{}{"subject": ticket.Subject}),
	)

	s.forwardTicket(&ticket, employeeLogin, subject, description, screenshots, sourceDomain)

	shots := s.loadScreenshots([]int64{ticket.IDHelpdeskTicket})
	r := toTicketResult(ticket, shots[ticket.IDHelpdeskTicket])
	return &r, nil
}

// ReceiveReply handles an incoming reply from the external helpdesk (HMAC-protected, no JWT).
func (s *Service) ReceiveReply(ticketID int64, gotHMAC string, body map[string]interface{}) (*TicketResult, error) {
	expected := strings.TrimSpace(s.cfg.HelpdeskReplyHMAC)
	if expected == "" {
		return nil, fmt.Errorf("HELPDESK_REPLY_HMAC not configured")
	}
	if !hmac.Equal([]byte(strings.TrimSpace(gotHMAC)), []byte(expected)) {
		return nil, fmt.Errorf("unauthorized")
	}

	var ticket genModel.HelpdeskTicket
	if err := s.db.Where("id_helpdesk_ticket = ?", ticketID).First(&ticket).Error; err != nil {
		return nil, fmt.Errorf("ticket not found")
	}

	// Accept {text,links,images,...} or {payload:{...},...}
	base := body
	if nested, ok := body["payload"].(map[string]interface{}); ok {
		base = nested
	}

	text := ""
	if v, ok := base["text"].(string); ok {
		text = strings.TrimSpace(v)
	}
	var links []string
	if l, ok := base["links"].([]interface{}); ok {
		for _, x := range l {
			if v := strings.TrimSpace(fmt.Sprintf("%v", x)); v != "" {
				links = append(links, v)
			}
		}
	}
	var images []string
	if l, ok := base["images"].([]interface{}); ok {
		for _, x := range l {
			if v := strings.TrimSpace(fmt.Sprintf("%v", x)); v != "" {
				images = append(images, v)
			}
		}
	}

	moderator, _ := body["moderator"].(map[string]interface{})
	mid, fn, ln := "", "", ""
	if moderator != nil {
		if v, ok := moderator["moderator_id"]; ok {
			mid = fmt.Sprintf("%v", v)
		}
		fn, _ = moderator["first_name"].(string)
		ln, _ = moderator["last_name"].(string)
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	header := fmt.Sprintf("\n\n---\n[Support reply %s", ts)
	if mid != "" {
		header += fmt.Sprintf(" by %s %s (id=%s)", strings.TrimSpace(fn), strings.TrimSpace(ln), mid)
	}
	header += "]\n"

	appendText := header
	if text != "" {
		appendText += text + "\n"
	}
	if len(links) > 0 {
		appendText += "\nLinks:\n" + strings.Join(links, "\n") + "\n"
	}
	ticket.Description = ticket.Description + appendText

	for _, u := range images {
		s.db.Create(&genModel.HelpdeskTicketScreenshot{TicketID: ticket.IDHelpdeskTicket, URL: u})
	}

	if closed, _ := body["closed"].(bool); closed {
		ticket.Status = "closed"
	}

	if err := s.db.Save(&ticket).Error; err != nil {
		return nil, err
	}

	shots := s.loadScreenshots([]int64{ticket.IDHelpdeskTicket})
	r := toTicketResult(ticket, shots[ticket.IDHelpdeskTicket])
	return &r, nil
}

// ── Private ───────────────────────────────────────────────────────────────────

func (s *Service) loadScreenshots(ids []int64) map[int64][]ScreenshotItem {
	result := map[int64][]ScreenshotItem{}
	if len(ids) == 0 {
		return result
	}
	var rows []genModel.HelpdeskTicketScreenshot
	s.db.Where("ticket_id IN ?", ids).Find(&rows)
	for _, r := range rows {
		result[r.TicketID] = append(result[r.TicketID], ScreenshotItem{
			IDHelpdeskTicketScreenshot: r.IDHelpdeskTicketScreenshot,
			URL:                        r.URL,
			CreatedAt:                  r.CreatedAt,
		})
	}
	return result
}

func (s *Service) forwardTicket(ticket *genModel.HelpdeskTicket, employeeLogin, subject, description string, screenshots []string, sourceDomain *string) {
	now := time.Now().UTC()
	ticket.ForwardedAt = &now

	forwardURL := s.cfg.HelpdeskForwardURL
	if forwardURL == "" {
		forwardURL = "https://console.devinsidercode.com/helpdesk/sighthub"
	}

	forwardHMAC := strings.TrimSpace(s.cfg.HelpdeskForwardHMAC)
	if forwardHMAC == "" {
		ticket.ForwardOk = false
		errMsg := "HELPDESK_FORWARD_HMAC is not configured"
		ticket.ForwardError = &errMsg
		s.db.Save(ticket)
		return
	}

	domain := ""
	if sourceDomain != nil {
		domain = *sourceDomain
	}
	payload := map[string]interface{}{
		"employee_login": employeeLogin,
		"subject":        subject,
		"description":    description,
		"screenshots":    screenshots,
		"source_domain":  domain,
		"ticket_id":      ticket.IDHelpdeskTicket,
	}
	bodyBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", forwardURL, bytes.NewReader(bodyBytes))
	if err != nil {
		ticket.ForwardOk = false
		errMsg := fmt.Sprintf("forward request build error: %v", err)
		ticket.ForwardError = &errMsg
		s.db.Save(ticket)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trusted-HMAC", forwardHMAC)
	req.Header.Set("X-Trusted-Service", "helpdesk_sighthub")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		ticket.ForwardOk = false
		errMsg := fmt.Sprintf("forward exception: %v", err)
		ticket.ForwardError = &errMsg
		s.db.Save(ticket)
		return
	}
	defer resp.Body.Close()

	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	ticket.ForwardOk = ok

	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		respBody = map[string]interface{}{"status_code": resp.StatusCode}
	}
	respJSON, _ := json.Marshal(respBody)
	respStr := string(respJSON)
	ticket.ForwardResponse = &respStr

	if !ok {
		errMsg := fmt.Sprintf("forward failed: status=%d", resp.StatusCode)
		ticket.ForwardError = &errMsg
	}
	s.db.Save(ticket)
}

func toTicketResult(t genModel.HelpdeskTicket, screenshots []ScreenshotItem) TicketResult {
	if screenshots == nil {
		screenshots = []ScreenshotItem{}
	}
	r := TicketResult{
		IDHelpdeskTicket: t.IDHelpdeskTicket,
		EmployeeLogin:    t.EmployeeLogin,
		Subject:          t.Subject,
		Description:      t.Description,
		SourceDomain:     t.SourceDomain,
		Status:           t.Status,
		CreatedAt:        t.CreatedAt.Format(time.RFC3339),
		ForwardOk:        t.ForwardOk,
		ForwardError:     t.ForwardError,
		Screenshots:      screenshots,
	}
	if t.ForwardedAt != nil {
		fwd := t.ForwardedAt.Format(time.RFC3339)
		r.ForwardedAt = &fwd
	}
	return r
}
