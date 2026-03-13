package ticket_service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	labTicketModel "sighthub-backend/internal/models/lab_ticket"
	"sighthub-backend/internal/models/notifications"
	invoiceModel "sighthub-backend/internal/models/invoices"
	locationModel "sighthub-backend/internal/models/location"
	patientModel "sighthub-backend/internal/models/patients"
	pkgComm "sighthub-backend/pkg/communication"
	pkgEmail "sighthub-backend/pkg/email"
	"sighthub-backend/pkg/scheduler"
)

const notifyDelay = time.Hour

type Service struct {
	db   *gorm.DB
	sched *scheduler.Scheduler
}

func New(db *gorm.DB, sched *scheduler.Scheduler) *Service {
	return &Service{db: db, sched: sched}
}

type NotifyResult struct {
	Message   string  `json:"message"`
	Ticket    string  `json:"ticket"`
	SendEmail bool    `json:"send_email"`
	SendSMS   bool    `json:"send_sms"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
}

// NotifyPatient schedules an email+SMS notification 1 hour from now.
// If a notification is already pending for this ticket it is replaced.
func (s *Service) NotifyPatient(ticketID int64) (*NotifyResult, error) {
	// Load ticket
	var ticket labTicketModel.LabTicket
	err := s.db.
		Preload("LabTicketStatus").
		First(&ticket, ticketID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ticket not found")
		}
		return nil, err
	}

	// Load patient
	var patient patientModel.Patient
	if err := s.db.First(&patient, ticket.PatientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	// Load invoice → location
	var invoice invoiceModel.Invoice
	var locationName string
	var locationID *int64
	if err := s.db.First(&invoice, ticket.InvoiceID).Error; err == nil {
		var loc locationModel.Location
		if err2 := s.db.First(&loc, invoice.LocationID).Error; err2 == nil {
			locationName = loc.FullName
			lid := int64(loc.IDLocation)
			locationID = &lid
		}
	}
	if locationName == "" {
		locationName = "Your Care Team"
	}

	// Resolve notify channels (default: both true)
	var setting notifications.NotifySetting
	sendEmail := true
	sendSMS := true
	if err := s.db.Where("action_name = ?", "ticket-status").First(&setting).Error; err == nil {
		sendEmail = setting.Email
		sendSMS = setting.SMS
	}

	statusText := "In progress"
	if ticket.LabTicketStatus != nil {
		statusText = ticket.LabTicketStatus.TicketStatus
	}
	patientName := fmt.Sprintf("%s %s", patient.FirstName, patient.LastName)
	orderNumber := ticket.NumberTicket
	orderTotal := ""
	if ticket.Amt != nil {
		orderTotal = *ticket.Amt
	}

	emailTemplate := pkgEmail.GetTemplateForCategory(s.db, "order")

	var emailTo *string
	var phoneTo *string
	if sendEmail && patient.Email != nil {
		emailTo = patient.Email
	}
	if sendSMS && patient.Phone != nil {
		phoneTo = patient.Phone
	}

	// Schedule goroutine (replaces celery apply_async countdown=3600)
	key := fmt.Sprintf("ticket_notify_%d", ticketID)
	s.sched.Schedule(key, notifyDelay, func() {
		if emailTo != nil {
			ctx := map[string]interface{}{
				"patient_name":      patientName,
				"organization_name": locationName,
				"order_number":      orderNumber,
				"order_items":       statusText,
				"order_total":       orderTotal,
			}
			subject := fmt.Sprintf("Order %s — %s", orderNumber, statusText)
			if err := pkgEmail.SendViaDB(s.db, *emailTo, subject, emailTemplate, ctx, locationID); err != nil {
				log.Printf("ticket notify email error (ticket %d): %v", ticketID, err)
			}
		}
		if phoneTo != nil {
			msg := fmt.Sprintf("Hi %s, order %s: %s. — %s", patientName, orderNumber, statusText, locationName)
			pkgComm.SendSMS(*phoneTo, msg)
		}
	})

	return &NotifyResult{
		Message:   fmt.Sprintf("Notification scheduled in 1 hour for patient %s", patientName),
		Ticket:    orderNumber,
		SendEmail: sendEmail,
		SendSMS:   sendSMS,
		Email:     emailTo,
		Phone:     phoneTo,
	}, nil
}
