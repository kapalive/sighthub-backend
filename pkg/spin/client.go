// Package spin implements an HTTP client for the iPOSpays SPIn REST API v2.
// Docs: https://app.theneo.io/dejavoo/spin/spin-rest-api-methods
package spin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	SandboxBaseURL    = "https://test.spinpos.net/spin"
	ProductionBaseURL = "https://api.spinpos.net"
)

// Client talks to the SPIn REST API.
type Client struct {
	BaseURL    string
	AuthKey    string
	TPN        string
	Timeout    time.Duration
	HTTPClient *http.Client
}

// NewClient creates a SPIn client. If timeout <= 0, defaults to 130s.
func NewClient(baseURL, authKey, tpn string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 130 * time.Second
	}
	return &Client{
		BaseURL: baseURL,
		AuthKey: authKey,
		TPN:     tpn,
		Timeout: timeout,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// ── Request / Response types ────────────────────────────────────────────────

// SaleRequest is sent to POST /v2/Payment/Sale.
type SaleRequest struct {
	Authkey          string  `json:"Authkey"`
	Tpn              string  `json:"Tpn"`
	Amount           float64 `json:"Amount"`
	PaymentType      string  `json:"PaymentType"`      // Credit, Debit, Card, Cash, userChoice
	ReferenceId      string  `json:"ReferenceId"`       // unique, max 50 chars
	InvoiceNumber    string  `json:"InvoiceNumber,omitempty"`
	CaptureSignature bool    `json:"CaptureSignature"`
	GetExtendedData  bool    `json:"GetExtendedData"`
	PrintReceipt     string  `json:"PrintReceipt,omitempty"` // No, Both, Merchant, Customer
	GetReceipt       string  `json:"GetReceipt,omitempty"`
}

// ReturnRequest is sent to POST /v2/Payment/Return.
type ReturnRequest struct {
	Authkey          string  `json:"Authkey"`
	Tpn              string  `json:"Tpn"`
	Amount           float64 `json:"Amount"`
	PaymentType      string  `json:"PaymentType"`
	ReferenceId      string  `json:"ReferenceId"`
	InvoiceNumber    string  `json:"InvoiceNumber,omitempty"`
	CaptureSignature bool    `json:"CaptureSignature"`
	GetExtendedData  bool    `json:"GetExtendedData"`
}

// VoidRequest is sent to POST /v2/Payment/Void.
type VoidRequest struct {
	Authkey          string  `json:"Authkey"`
	Tpn              string  `json:"Tpn"`
	Amount           float64 `json:"Amount"`
	ReferenceId      string  `json:"ReferenceId"` // original transaction ref
	GetExtendedData  bool    `json:"GetExtendedData"`
}

// StatusRequest is sent to POST /v2/Payment/Status.
type StatusRequest struct {
	Authkey     string `json:"Authkey"`
	Tpn         string `json:"Tpn"`
	ReferenceId string `json:"ReferenceId"`
}

// CardData returned in sale/return responses.
type CardData struct {
	Last4       string `json:"Last4"`
	First4      string `json:"First4"`
	CardType    string `json:"CardType"`
	EntryMethod string `json:"EntryMethod"` // Chip, Swipe, Contactless, Manual
	BIN         string `json:"BIN"`
	Expiration  string `json:"Expiration"`
}

// Amounts in the response.
type Amounts struct {
	TotalAmount float64 `json:"TotalAmount"`
	Amount      float64 `json:"Amount"`
	TipAmount   float64 `json:"TipAmount"`
	FeeAmount   float64 `json:"FeeAmount"`
	TaxAmount   float64 `json:"TaxAmount"`
}

// EMVData in the response.
type EMVData struct {
	AID string `json:"AID"`
	TVR string `json:"TVR"`
	TSI string `json:"TSI"`
	IAD string `json:"IAD"`
	ARC string `json:"ARC"`
}

// Response is the general SPIn REST API response.
type Response struct {
	ResultCode        int       `json:"ResultCode"`  // 0=OK, 1=Terminal Error, 2=API Error
	StatusCode        string    `json:"StatusCode"`  // 0000=Approved
	Message           string    `json:"Message"`
	DetailedMessage   string    `json:"DetailedMessage"`
	HostResponseCode  string    `json:"HostResponseCode"`
	HostResponseMsg   string    `json:"HostResponseMessage"`
	AuthCode          string    `json:"AuthCode"`
	ReferenceId       string    `json:"ReferenceId"`
	SerialNumber      string    `json:"SerialNumber"`
	BatchNumber       string    `json:"BatchNumber"`
	TransactionNumber string    `json:"TransactionNumber"`
	Token             string    `json:"Token"`
	RRN               string    `json:"RRN"`
	CardData          *CardData `json:"CardData"`
	Amounts           *Amounts  `json:"Amounts"`
	EMVData           *EMVData  `json:"EMVData"`

	// Raw JSON for storage
	RawJSON string `json:"-"`
}

// IsApproved returns true when the transaction was approved.
func (r *Response) IsApproved() bool {
	return r.ResultCode == 0 && r.StatusCode == "0000"
}

// IsDeclined returns true for a decline (terminal-side).
func (r *Response) IsDeclined() bool {
	return r.ResultCode == 1
}

// ── API methods ─────────────────────────────────────────────────────────────

// Sale initiates a POS sale transaction on the terminal.
func (c *Client) Sale(amount float64, refID, invoiceNum string) (*Response, error) {
	req := SaleRequest{
		Authkey:          c.AuthKey,
		Tpn:              c.TPN,
		Amount:           amount,
		PaymentType:      "Credit",
		ReferenceId:      refID,
		InvoiceNumber:    invoiceNum,
		CaptureSignature: false,
		GetExtendedData:  true,
		PrintReceipt:     "No",
		GetReceipt:       "Both",
	}
	return c.post("/v2/Payment/Sale", req)
}

// Return processes a refund on the terminal.
func (c *Client) Return(amount float64, refID, invoiceNum string) (*Response, error) {
	req := ReturnRequest{
		Authkey:          c.AuthKey,
		Tpn:              c.TPN,
		Amount:           amount,
		PaymentType:      "Credit",
		ReferenceId:      refID,
		InvoiceNumber:    invoiceNum,
		CaptureSignature: false,
		GetExtendedData:  true,
	}
	return c.post("/v2/Payment/Return", req)
}

// Void cancels a pre-settlement transaction.
func (c *Client) Void(amount float64, refID string) (*Response, error) {
	req := VoidRequest{
		Authkey:         c.AuthKey,
		Tpn:             c.TPN,
		Amount:          amount,
		ReferenceId:     refID,
		GetExtendedData: true,
	}
	return c.post("/v2/Payment/Void", req)
}

// Status checks the status of a pending transaction.
func (c *Client) Status(refID string) (*Response, error) {
	req := StatusRequest{
		Authkey:     c.AuthKey,
		Tpn:         c.TPN,
		ReferenceId: refID,
	}
	return c.post("/v2/Payment/Status", req)
}

// TerminalStatus checks if the terminal is online. GET /v2/Common/TerminalStatus
func (c *Client) TerminalStatus() (*Response, error) {
	url := c.BaseURL + "/v2/Common/TerminalStatus?Authkey=" + c.AuthKey + "&Tpn=" + c.TPN
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("spin: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	return c.doRequest(httpReq)
}

// ── internal ────────────────────────────────────────────────────────────────

func (c *Client) post(path string, body interface{}) (*Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("spin: marshal request: %w", err)
	}

	url := c.BaseURL + path
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("spin: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(httpReq)
	if resp != nil {
		resp.RawJSON = string(jsonBody) // store request JSON for audit
	}
	return resp, err
}

func (c *Client) doRequest(req *http.Request) (*Response, error) {
	httpResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("spin: http request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("spin: read response: %w", err)
	}

	var spinResp Response
	if err := json.Unmarshal(respBody, &spinResp); err != nil {
		return nil, fmt.Errorf("spin: unmarshal response (status %d): %s", httpResp.StatusCode, string(respBody))
	}
	spinResp.RawJSON = string(respBody)

	return &spinResp, nil
}
