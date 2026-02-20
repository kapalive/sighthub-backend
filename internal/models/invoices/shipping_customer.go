// internal/models/invoices/shipping_customer.go
package invoices

import (
	"fmt"
	"time"

	"sighthub-backend/internal/models/interfaces"
)

// ShippingCustomer ↔ table: shipping_customer
type ShippingCustomer struct {
	IDShippingCustomer int64      `gorm:"column:id_shipping_customer;primaryKey;autoIncrement" json:"id_shipping_customer"`
	InvoiceID          int64      `gorm:"column:invoice_id"                                   json:"invoice_id"`
	PatientID          *int64     `gorm:"column:patient_id"                                    json:"patient_id,omitempty"`
	ShippingServicesID *int64     `gorm:"column:shipping_services_id"                          json:"shipping_services_id,omitempty"`
	ShippingTrackerID  *int64     `gorm:"column:shipping_tracker_id"                           json:"shipping_tracker_id,omitempty"`
	ShipmentTypeID     *int64     `gorm:"column:shipment_type_id"                              json:"shipment_type_id,omitempty"`
	DateDispatch       *time.Time `gorm:"column:date_dispatch;type:timestamptz"                json:"date_dispatch,omitempty"`

	// Invoice в том же пакете — можно держать прямую связь
	Invoice *Invoice `gorm:"foreignKey:InvoiceID;references:IDInvoice" json:"invoice,omitempty"`

	// Пациента держим через интерфейс (без прямой зависимости на пакет patients)
	// Поле не маппится в БД, только для временного хранения, если нужно отдать в JSON.
	PatientResolved map[string]interface{} `gorm:"-" json:"-"`
}

func (ShippingCustomer) TableName() string { return "shipping_customer" }

func (s *ShippingCustomer) String() string {
	return fmt.Sprintf("<ShippingCustomer %d>", s.IDShippingCustomer)
}

func (s *ShippingCustomer) ToMap() map[string]interface{} {
	var dt *string
	if s.DateDispatch != nil {
		iso := s.DateDispatch.UTC().Format(time.RFC3339)
		dt = &iso
	}
	m := map[string]interface{}{
		"id_shipping_customer": s.IDShippingCustomer,
		"invoice_id":           s.InvoiceID,
		"patient_id":           s.PatientID,
		"shipping_services_id": s.ShippingServicesID,
		"shipping_tracker_id":  s.ShippingTrackerID,
		"shipment_type_id":     s.ShipmentTypeID,
		"date_dispatch":        dt,
	}

	// Вложенный invoice — если прелоадили
	if s.Invoice != nil {
		m["invoice"] = s.Invoice.ToMap()
	}

	// Вложенный patient — если заранее резолвили через интерфейс
	if s.PatientResolved != nil {
		m["patient"] = s.PatientResolved
	}

	// Остальные (shipping_service / shipping_tracker / shipment_type)
	// можно резолвить на уровне сервиса по их *_id при необходимости.
	return m
}

// GetPatient — получить пациента через интерфейс и (опционально) сохранить во временное поле
func (s *ShippingCustomer) GetPatient(resolver interfaces.PatientInterface, store bool) (map[string]interface{}, error) {
	if s.PatientID == nil {
		return nil, nil
	}
	data, err := resolver.GetPatientByID(*s.PatientID)
	if err != nil {
		return nil, err
	}
	if store {
		s.PatientResolved = data
	}
	return data, nil
}
