package zeiss_service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	labTicketModel "sighthub-backend/internal/models/lab_ticket"
	patientModel "sighthub-backend/internal/models/patients"
)

// ─── Result structs ─────────────────────────────────────────────────────────

type ZeissOrderResult struct {
	OrderID          string `json:"order_id"`
	Status           string `json:"status"`
	Message          string `json:"message,omitempty"`
	ConfirmedOrderID string `json:"confirmed_order_id,omitempty"`
}

type ZeissOrderStatusResult struct {
	OrderID           string `json:"order_id"`
	SalesOrderID      string `json:"sales_order_id,omitempty"`
	OrderStatus       string `json:"order_status"`
	ConfirmedDate     string `json:"confirmed_date,omitempty"`
	EstimatedDelivery string `json:"estimated_delivery,omitempty"`
}

// ─── XML response structs ───────────────────────────────────────────────────

type b2bOpticResponse struct {
	XMLName xml.Name            `xml:"b2bOptic"`
	Header  *b2bRespHeader      `xml:"header"`
	Items   *b2bRespItems       `xml:"items"`
	Errors  *b2bRespErrors      `xml:"errors"`
}

type b2bRespHeader struct {
	MsgType string `xml:"msgType,attr"`
	OrderID string `xml:"customersOrderId"`
}

type b2bRespItems struct {
	Items []b2bRespItem `xml:"item"`
}

type b2bRespItem struct {
	ReferenceNo      string `xml:"referenceNo"`
	ConfirmedOrderID string `xml:"confirmedOrderId"`
}

type b2bRespErrors struct {
	Errors []b2bRespError `xml:"error"`
}

type b2bRespError struct {
	Code    string `xml:"code,attr"`
	Message string `xml:",chardata"`
}

// ─── XML status response structs ────────────────────────────────────────────

type b2bOpticInfoResponse struct {
	XMLName xml.Name              `xml:"b2bOpticInfo"`
	Items   *b2bInfoItems         `xml:"items"`
}

type b2bInfoItems struct {
	Items []b2bInfoItem `xml:"item"`
}

type b2bInfoItem struct {
	ReferenceNo           string `xml:"referenceNo"`
	ConfirmedOrderID      string `xml:"confirmedOrderId"`
	SalesOrderID          string `xml:"salesOrderId"`
	ConfirmedOrderDate    string `xml:"confirmedOrderDate"`
	EstimatedDeliveryDate string `xml:"estimatedDeliveryDate"`
	OrderStatus           string `xml:"orderStatus"`
}

// ─── PlaceZeissOrder ────────────────────────────────────────────────────────

func (s *CatalogService) PlaceZeissOrder(employeeID int64, ticketID int64) (*ZeissOrderResult, error) {
	// a) Get token
	token, err := s.auth.GetToken(employeeID)
	if err != nil {
		return nil, fmt.Errorf("zeiss order: %w", err)
	}

	// b) Get customer number
	status := s.auth.GetAuthStatus(employeeID)
	if status.CustomerNumber == nil || *status.CustomerNumber == "" {
		return nil, fmt.Errorf("zeiss order: customer number not available — re-authenticate with Zeiss")
	}
	customerNumber := *status.CustomerNumber
	custID := customerNumber // use full customer number (e.g. "632541.211395")

	// c) Load ticket with preloads
	var ticket labTicketModel.LabTicket
	if err := s.db.
		Preload("Powers").
		Preload("Lens").
		Preload("Frame").
		Preload("Lab").
		First(&ticket, ticketID).Error; err != nil {
		return nil, fmt.Errorf("zeiss order: ticket not found: %w", err)
	}

	// Check if already ordered
	if ticket.VwOrderID != nil && *ticket.VwOrderID != "" {
		return &ZeissOrderResult{
			OrderID: *ticket.VwOrderID,
			Status:  "already_submitted",
			Message: "Order already submitted",
		}, fmt.Errorf("zeiss order: ticket %s already submitted as order %s", ticket.NumberTicket, *ticket.VwOrderID)
	}

	// Verify lab is Zeiss
	if ticket.LabID == nil || *ticket.LabID != ZeissVendorID {
		return nil, fmt.Errorf("zeiss order: ticket lab is not CARL ZEISS")
	}

	// d) Load patient
	var patient patientModel.Patient
	if err := s.db.First(&patient, ticket.PatientID).Error; err != nil {
		return nil, fmt.Errorf("zeiss order: patient not found: %w", err)
	}

	// e) Get lens commercial code
	lens := ticket.Lens
	if lens == nil {
		return nil, fmt.Errorf("zeiss order: no lens data on ticket")
	}
	// Auto-sync from invoice if needed
	if lens.VwDesignCode == nil || *lens.VwDesignCode == "" {
		s.trySyncLensFromInvoice(ticket.InvoiceID, lens)
	}
	if lens.VwDesignCode == nil || *lens.VwDesignCode == "" {
		return nil, fmt.Errorf("zeiss order: lens commercial code (vw_design_code) is missing")
	}
	commercialCode := *lens.VwDesignCode

	// f) Get treatment codes from invoice, classify as COATING vs COLOR via PCAT
	var coatingCode, colorCode string
	if ticket.InvoiceID > 0 {
		type treatRow struct {
			VwCode string
		}
		var treats []treatRow
		s.db.Raw(`
			SELECT lt.vw_code
			FROM invoice_item_sale iis
			JOIN lens_treatments lt ON lt.id_lens_treatments = iis.item_id AND iis.item_type = 'Treatment'
			WHERE iis.invoice_id = ? AND lt.source = 'zeiss_only' AND lt.vw_code IS NOT NULL AND lt.vw_code != ''
		`, ticket.InvoiceID).Scan(&treats)

		if len(treats) > 0 {
			// Get allowed treatments with types from PCAT
			allowed, _ := s.GetAllowedTreatments(employeeID, commercialCode, customerNumber)
			typeMap := make(map[string]string) // vw_code → type
			for _, a := range allowed {
				typeMap[a.VwCode] = a.Type
			}
			for _, t := range treats {
				switch typeMap[t.VwCode] {
				case "COLOR":
					if colorCode == "" {
						colorCode = t.VwCode
					}
				default: // COATING or unknown
					if coatingCode == "" {
						coatingCode = t.VwCode
					}
				}
			}
		}
	}

	// g) Determine order mode
	orderMode := "complete"
	msgType := "ORDER"
	if os.Getenv("APP_ENV") == "development" {
		orderMode = "incomplete"
		msgType = "CALCULATION"
	}

	// Build b2bOptic XML
	xmlBody := buildOrderXML(
		custID,
		ticket.NumberTicket,
		&patient,
		ticket.Powers,
		ticket.Frame,
		commercialCode,
		coatingCode,
		colorCode,
		msgType,
	)

	log.Printf("[zeiss] placing %s order for ticket %s, customer %s", orderMode, ticket.NumberTicket, custID)

	// h) POST order
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/public/api/b2boptic/v2/orders/%s?versionIn=1.6.5&infoOut=0.3", s.auth.APIBase(), orderMode)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(xmlBody))
	if err != nil {
		return nil, fmt.Errorf("zeiss order: create request: %w", err)
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", s.auth.OrderSubKey())
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/xml")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zeiss order: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("zeiss order: read response: %w", err)
	}

	log.Printf("[zeiss] order response status=%d body=%s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != 202 {
		return nil, fmt.Errorf("zeiss order: API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// i) Parse response XML
	result, err := parseOrderResponse(respBody)
	if err != nil {
		return nil, fmt.Errorf("zeiss order: parse response: %w", err)
	}

	// j) Save order ID to ticket
	if result.OrderID != "" || result.ConfirmedOrderID != "" {
		orderIDToSave := result.ConfirmedOrderID
		if orderIDToSave == "" {
			orderIDToSave = result.OrderID
		}
		s.db.Model(&labTicketModel.LabTicket{}).
			Where("id_lab_ticket = ?", ticketID).
			Update("vw_order_id", orderIDToSave)
		result.OrderID = orderIDToSave
	}

	return result, nil
}

// ─── GetZeissOrderStatus ────────────────────────────────────────────────────

func (s *CatalogService) GetZeissOrderStatus(employeeID int64, orderID string, customerNumber string) (*ZeissOrderStatusResult, error) {
	// a) Get token
	token, err := s.auth.GetToken(employeeID)
	if err != nil {
		return nil, fmt.Errorf("zeiss status: %w", err)
	}

	custID := customerNumber // full customer number for API

	// b) GET /orders/status
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/public/api/b2boptic/v2/orders/status?infoOut=0.3&originator=%s&orderId=%s",
		s.auth.APIBase(), custID, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("zeiss status: create request: %w", err)
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", s.auth.OrderSubKey())
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zeiss status: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("zeiss status: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zeiss status: API returned status %d: %s", resp.StatusCode, string(body))
	}

	// c) Parse XML response
	var info b2bOpticInfoResponse
	if err := xml.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("zeiss status: parse response: %w", err)
	}

	if info.Items == nil || len(info.Items.Items) == 0 {
		return &ZeissOrderStatusResult{
			OrderID:     orderID,
			OrderStatus: "UNKNOWN",
		}, nil
	}

	item := info.Items.Items[0]
	return &ZeissOrderStatusResult{
		OrderID:           item.ConfirmedOrderID,
		SalesOrderID:      item.SalesOrderID,
		OrderStatus:       item.OrderStatus,
		ConfirmedDate:     item.ConfirmedOrderDate,
		EstimatedDelivery: item.EstimatedDeliveryDate,
	}, nil
}

// GetZeissAuthStatus exposes the auth status for use by handlers.
func (s *CatalogService) GetZeissAuthStatus(employeeID int64) AuthStatus {
	return s.auth.GetAuthStatus(employeeID)
}

// ─── helpers ────────────────────────────────────────────────────────────────

// zeissShortCustomerNumber extracts the short form from "632541.211395" → "0000211395"
// Takes the part after the dot, pads to 10 digits with leading zeros.
func zeissShortCustomerNumber(full string) string {
	parts := strings.SplitN(full, ".", 2)
	short := full
	if len(parts) == 2 {
		short = parts[1]
	}
	// Pad to 10 digits
	for len(short) < 10 {
		short = "0" + short
	}
	return short
}

// buildOrderXML constructs the b2bOptic XML for a Zeiss order.
func buildOrderXML(
	custID string,
	ticketNumber string,
	patient *patientModel.Patient,
	powers *labTicketModel.LabTicketPowers,
	frame *labTicketModel.LabTicketFrame,
	commercialCode string,
	coatingCode string,
	colorCode string,
	msgType string,
) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n<b2bOptic>")

	// ── Header ──
	fmt.Fprintf(&b, "\n  <header msgType=\"%s\" msgState=\"NEW\">", msgType)
	fmt.Fprintf(&b, "\n    <customersOrderId>%s</customersOrderId>", xmlEscape(ticketNumber))
	b.WriteString("\n    <distributorsOrderId>-</distributorsOrderId>")
	fmt.Fprintf(&b, "\n    <timeStamps><dateTime step=\"CREATE\">%s</dateTime></timeStamps>", time.Now().UTC().Format("2006-01-02T15:04:05.000Z"))
	fmt.Fprintf(&b, "\n    <orderParties role=\"ORIGINATOR\"><id memberShipID=\"1\">%s</id><name>SightHub</name></orderParties>", custID)
	fmt.Fprintf(&b, "\n    <orderParties role=\"SHIPTO\"><id memberShipID=\"1\">%s</id><name>SightHub</name></orderParties>", custID)
	b.WriteString("\n    <software typeOf=\"ORIGINATOR\"><name>SightHub PMS</name><version>1.0</version></software>")
	fmt.Fprintf(&b, "\n    <portalOrderId>%s</portalOrderId>", xmlEscape(ticketNumber))
	b.WriteString("\n  </header>")

	// ── Items ──
	b.WriteString("\n  <items>")
	b.WriteString("\n    <item>")
	fmt.Fprintf(&b, "\n      <parties role=\"ORIGINATOR\"><id memberShipID=\"1\">%s</id><name>SightHub</name></parties>", custID)
	fmt.Fprintf(&b, "\n      <parties role=\"SHIPTO\"><id memberShipID=\"1\">%s</id><name>SightHub</name></parties>", custID)
	fmt.Fprintf(&b, "\n      <referenceNo>%s</referenceNo>", xmlEscape(ticketNumber))

	b.WriteString("\n      <manufacturer/>")
	b.WriteString("\n      <pair>")

	// Patient
	b.WriteString("\n        <patient>")
	fmt.Fprintf(&b, "\n          <id memberShipID=\"1\">%s</id>", custID)
	fmt.Fprintf(&b, "\n          <name>%s, %s</name>", xmlEscape(patient.LastName), xmlEscape(patient.FirstName))
	b.WriteString("\n        </patient>")

	// Right lens
	buildLensXML(&b, "RIGHT", commercialCode, coatingCode, colorCode, powers, true)

	// Left lens
	buildLensXML(&b, "LEFT", commercialCode, coatingCode, colorCode, powers, false)

	// Frame
	if frame != nil {
		b.WriteString("\n        <frame quantity=\"0\">")
		b.WriteString("\n          <material>SPECIAL</material>")
		if frame.SizeLensWidth != nil && *frame.SizeLensWidth != "" {
			fmt.Fprintf(&b, "\n          <boxWidth>%s</boxWidth>", *frame.SizeLensWidth)
		}
		if frame.BValue != nil {
			fmt.Fprintf(&b, "\n          <boxHeight>%d</boxHeight>", *frame.BValue)
		}
		if frame.SizeBridgeWidth != nil && *frame.SizeBridgeWidth != "" {
			fmt.Fprintf(&b, "\n          <distanceBetweenLenses>%s</distanceBetweenLenses>", *frame.SizeBridgeWidth)
		}
		if frame.Panto != nil {
			fmt.Fprintf(&b, "\n          <pantoscopicAngle dimension=\"DEG\">%s</pantoscopicAngle>", formatFloat(*frame.Panto))
		}
		if frame.WrapAngle != nil {
			fmt.Fprintf(&b, "\n          <frameBowAngle>%s</frameBowAngle>", formatFloat(*frame.WrapAngle))
		}
		b.WriteString("\n        </frame>")
	}

	b.WriteString("\n      </pair>")
	b.WriteString("\n    </item>")
	b.WriteString("\n  </items>")
	b.WriteString("\n</b2bOptic>")

	return b.String()
}

// buildLensXML writes a <lens> element for one side.
func buildLensXML(b *strings.Builder, side string, commercialCode string, coatingCode string, colorCode string, powers *labTicketModel.LabTicketPowers, isRight bool) {
	fmt.Fprintf(b, "\n        <lens side=\"%s\" quantity=\"1\">", side)
	fmt.Fprintf(b, "\n          <commercialCode>%s</commercialCode>", xmlEscape(commercialCode))

	// rxData
	if powers != nil {
		var sph, cyl, axis *string
		var add *float64
		var hPrism *float64
		var vPrism *float64
		var dt, segHD, bvd *string

		if isRight {
			sph = powers.ODSph
			cyl = powers.ODCyl
			axis = powers.ODAxis
			add = powers.ODAdd
			hPrism = powers.ODHPrism
			vPrism = powers.ODVPrism
			dt = powers.ODDT
			segHD = powers.ODSegHD
			bvd = powers.ODBVD
		} else {
			sph = powers.OSSph
			cyl = powers.OSCyl
			axis = powers.OSAxis
			add = powers.OSAdd
			hPrism = powers.OSHPrism
			vPrism = powers.OSVPrism
			dt = powers.OSDT
			segHD = powers.OSSegHD
			bvd = powers.OSBVD
		}

		b.WriteString("\n          <rxData>")
		if sph != nil && *sph != "" {
			fmt.Fprintf(b, "\n            <sphere>%s</sphere>", formatRxValue(*sph))
		}
		if cyl != nil && *cyl != "" {
			fmt.Fprintf(b, "\n            <cylinder>")
			fmt.Fprintf(b, "\n              <power>%s</power>", formatRxValue(*cyl))
			if axis != nil && *axis != "" {
				fmt.Fprintf(b, "\n              <axis>%s</axis>", formatRxValue(*axis))
			}
			fmt.Fprintf(b, "\n            </cylinder>")
		}
		if add != nil {
			fmt.Fprintf(b, "\n            <addition>%s</addition>", formatFloat(*add))
		}
		if (hPrism != nil && *hPrism != 0) || (vPrism != nil && *vPrism != 0) {
			b.WriteString("\n            <prism>")
			if hPrism != nil && *hPrism != 0 {
				fmt.Fprintf(b, "\n              <horizontal>%s</horizontal>", formatFloat(*hPrism))
			}
			if vPrism != nil && *vPrism != 0 {
				fmt.Fprintf(b, "\n              <vertical>%s</vertical>", formatFloat(*vPrism))
			}
			b.WriteString("\n            </prism>")
		}
		b.WriteString("\n          </rxData>")

		// Coating (one ANTIREFLEX + one COLOR max)
		if coatingCode != "" {
			fmt.Fprintf(b, "\n          <coating coatingType=\"ANTIREFLEX\"><commercialCode>%s</commercialCode></coating>", xmlEscape(coatingCode))
		}
		if colorCode != "" {
			fmt.Fprintf(b, "\n          <coating coatingType=\"COLOR\"><commercialCode>%s</commercialCode></coating>", xmlEscape(colorCode))
		}

		// Centration
		hasCentration := (dt != nil && *dt != "") || (segHD != nil && *segHD != "") || (bvd != nil && *bvd != "")
		if hasCentration {
			b.WriteString("\n          <centration>")
			if dt != nil && *dt != "" {
				fmt.Fprintf(b, "\n            <monocularCentrationDistance reference=\"FAR\">%s</monocularCentrationDistance>", formatRxValue(*dt))
			}
			if segHD != nil && *segHD != "" {
				fmt.Fprintf(b, "\n            <height reference=\"FAR\" referenceHeight=\"OVERBOX\">%s</height>", formatRxValue(*segHD))
			}
			if bvd != nil && *bvd != "" {
				fmt.Fprintf(b, "\n            <backVertexDistance rxDataNeedRecalculation=\"false\">%s</backVertexDistance>", formatRxValue(*bvd))
			}
			b.WriteString("\n          </centration>")
		}
	}

	// Geometry — fixed defaults
	b.WriteString("\n          <geometry>")
	b.WriteString("\n            <diameter>")
	b.WriteString("\n              <physical>65</physical>")
	b.WriteString("\n              <optical>70</optical>")
	b.WriteString("\n            </diameter>")
	b.WriteString("\n          </geometry>")

	b.WriteString("\n        </lens>")
}

// formatRxValue converts a string Rx value to a clean number string.
func formatRxValue(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "0"
	}
	// Try to parse as float to normalize
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return formatFloat(f)
	}
	return xmlEscape(s)
}

// formatFloat formats a float64 without unnecessary trailing zeros.
func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	// Trim trailing zeros after decimal point but keep at least one decimal
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// xmlEscape escapes special XML characters.
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// parseOrderResponse parses the b2bOptic XML response from Zeiss.
func parseOrderResponse(body []byte) (*ZeissOrderResult, error) {
	var resp b2bOpticResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		// If we can't parse as b2bOptic, try b2bOpticInfo
		var info b2bOpticInfoResponse
		if err2 := xml.Unmarshal(body, &info); err2 == nil && info.Items != nil && len(info.Items.Items) > 0 {
			item := info.Items.Items[0]
			return &ZeissOrderResult{
				OrderID:          item.ConfirmedOrderID,
				ConfirmedOrderID: item.ConfirmedOrderID,
				Status:           item.OrderStatus,
			}, nil
		}
		return nil, fmt.Errorf("invalid XML response: %w", err)
	}

	result := &ZeissOrderResult{
		Status: "submitted",
	}

	// Check for errors in response
	if resp.Errors != nil && len(resp.Errors.Errors) > 0 {
		var msgs []string
		for _, e := range resp.Errors.Errors {
			msgs = append(msgs, e.Message)
		}
		result.Status = "error"
		result.Message = strings.Join(msgs, "; ")
		return result, fmt.Errorf("zeiss order errors: %s", result.Message)
	}

	// Extract order ID from header
	if resp.Header != nil && resp.Header.OrderID != "" {
		result.OrderID = resp.Header.OrderID
	}

	// Extract confirmed order ID from items
	if resp.Items != nil && len(resp.Items.Items) > 0 {
		item := resp.Items.Items[0]
		if item.ConfirmedOrderID != "" {
			result.ConfirmedOrderID = item.ConfirmedOrderID
			result.OrderID = item.ConfirmedOrderID
		}
		if item.ReferenceNo != "" && result.OrderID == "" {
			result.OrderID = item.ReferenceNo
		}
	}

	result.Message = "Order submitted successfully"
	return result, nil
}
