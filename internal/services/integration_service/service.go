package integration_service

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	intModel "sighthub-backend/internal/models/integrations"

	"gorm.io/gorm"
)

const (
	vwAPIUser          = "qademo"
	vwAPIPass          = "xNH3L7K1wl9lezUZQwdA"
	vwTokenURL         = "https://regauth.visionweb.com/connect/token"
	vwUserAccountsURL  = "https://www.visionwebqa.com/services/services/UserAccountsService"
	vwRefID            = "ROSIGHTHUB"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) DB() *gorm.DB { return s.db }

func (s *Service) GetIntegration(code string, locationID int64) (map[string]interface{}, error) {
	var company intModel.IntegrationCompany
	if err := s.db.Where("code = ?", code).First(&company).Error; err != nil {
		return nil, fmt.Errorf("integration provider '%s' not found", code)
	}

	var config intModel.IntegrationConfig
	err := s.db.Where("integration_company_id = ? AND location_id = ?", company.IDIntegrationCompany, locationID).First(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return map[string]interface{}{
			"provider":    company.Name,
			"code":        company.Code,
			"location_id": locationID,
			"username":    "",
			"connected":   false,
			"last_check":  nil,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"provider":    company.Name,
		"code":        company.Code,
		"location_id": config.LocationID,
		"username":    config.Username,
		"connected":   config.Connected,
		"last_check":  config.LastCheck,
	}, nil
}

func (s *Service) SetIntegration(code string, locationID int64, username, password string) (map[string]interface{}, error) {
	var company intModel.IntegrationCompany
	if err := s.db.Where("code = ?", code).First(&company).Error; err != nil {
		return nil, fmt.Errorf("integration provider '%s' not found", code)
	}

	connected := false
	var checkErr string
	var labs []VWLabInfo

	switch code {
	case "vw":
		vwLabs, err := s.GetVisionWebLabs(username, password)
		if err != nil {
			checkErr = err.Error()
		} else {
			connected = true
			labs = vwLabs
		}
	case "zeiss":
		connected = false
	}

	now := time.Now()

	var config intModel.IntegrationConfig
	err := s.db.Where("integration_company_id = ? AND location_id = ?", company.IDIntegrationCompany, locationID).First(&config).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		config = intModel.IntegrationConfig{
			IntegrationCompanyID: company.IDIntegrationCompany,
			LocationID:           locationID,
			Username:             username,
			Password:             password,
			Connected:            connected,
			LastCheck:            &now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := s.db.Create(&config).Error; err != nil {
			return nil, fmt.Errorf("failed to save: %w", err)
		}
	} else if err == nil {
		config.Username = username
		config.Password = password
		config.Connected = connected
		config.LastCheck = &now
		config.UpdatedAt = now
		if err := s.db.Save(&config).Error; err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}
	} else {
		return nil, err
	}

	result := map[string]interface{}{
		"provider":    company.Name,
		"code":        company.Code,
		"location_id": locationID,
		"username":    username,
		"connected":   connected,
		"last_check":  now,
	}

	if !connected && checkErr != "" {
		result["error"] = checkErr
	}
	if labs != nil {
		result["labs"] = labs
	}

	return result, nil
}

// ─── VisionWeb SOAP UserAccountsService ──────────────────────────────────────

type vwUserProfile struct {
	XMLName  xml.Name       `xml:"VW_USER_PROFILE"`
	Login    vwLogin        `xml:"LOGIN"`
	CBU      vwCBU          `xml:"CBU"`
	Errors   vwErrorMessages `xml:"ERROR_MESSAGES"`
}

type vwLogin struct {
	Name string `xml:"Name,attr"`
}

type vwCBU struct {
	Suppliers []vwSupplier `xml:"SUPPLIERS>SUPPLIER"`
}

type vwSupplier struct {
	Name      string              `xml:"Name,attr"`
	Locations []vwSupplierLocation `xml:"SUPPLIER_LOCATIONS>SUPPLIER_LOCATION"`
}

type vwSupplierLocation struct {
	ID       string       `xml:"Id,attr"`
	Name     string       `xml:"Name,attr"`
	Accounts []vwAccount  `xml:"ACCOUNTS>ACCOUNT"`
}

type vwAccount struct {
	Billing  vwBillShip `xml:"BILLING_ACCOUNT"`
	Shipping vwBillShip `xml:"SHIPPING_ACCOUNT"`
}

type vwBillShip struct {
	Number string `xml:"Number,attr"`
}

type vwErrorMessages struct {
	Messages []vwErrorMsg `xml:"ERROR_MESSAGE"`
}

type vwErrorMsg struct {
	ID      string `xml:"ID"`
	Message string `xml:"MESSAGE"`
}

type VWLabInfo struct {
	SupplierName string `json:"supplier_name"`
	SloID        string `json:"slo_id"`
	LocationName string `json:"location_name"`
	BillAccount  string `json:"bill_account"`
	ShipAccount  string `json:"ship_account"`
}

func (s *Service) checkVisionWebAuth(username, password string) (bool, error) {
	_, err := s.GetVisionWebLabs(username, password)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) GetVisionWebLabs(username, password string) ([]VWLabInfo, error) {
	soapBody := fmt.Sprintf(`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:web="http://webservices.login.visionweb.com">
   <soapenv:Header/>
   <soapenv:Body>
      <web:getUserProfileByLogin>
         <web:username>%s</web:username>
         <web:password>%s</web:password>
         <web:refid>%s</web:refid>
      </web:getUserProfileByLogin>
   </soapenv:Body>
</soapenv:Envelope>`, username, password, vwRefID)

	req, _ := http.NewRequest("POST", vwUserAccountsURL, strings.NewReader(soapBody))
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"getUserProfileByLogin"`)

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("VisionWeb connection failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	bodyStr := string(respBody)

	// Extract VW_USER_PROFILE from SOAP response (HTML-encoded inside SOAP)
	// The response contains HTML entities that need to be unescaped
	start := strings.Index(bodyStr, "&lt;VW_USER_PROFILE")
	if start < 0 {
		start = strings.Index(bodyStr, "<VW_USER_PROFILE")
	}

	var profileXML string
	if start >= 0 && strings.Contains(bodyStr, "&lt;") {
		// HTML-encoded XML inside SOAP
		end := strings.Index(bodyStr[start:], "</p612:getUserProfileByLoginReturn>")
		if end < 0 {
			end = strings.Index(bodyStr[start:], "</")
			if end < 0 {
				end = len(bodyStr) - start
			}
		}
		raw := bodyStr[start : start+end]
		// Unescape HTML entities
		profileXML = strings.ReplaceAll(raw, "&lt;", "<")
		profileXML = strings.ReplaceAll(profileXML, "&gt;", ">")
		profileXML = strings.ReplaceAll(profileXML, "&amp;", "&")
		profileXML = strings.ReplaceAll(profileXML, "&quot;", `"`)
		profileXML = strings.ReplaceAll(profileXML, "&apos;", "'")
	} else if start >= 0 {
		end := strings.Index(bodyStr, "</VW_USER_PROFILE>")
		if end >= 0 {
			profileXML = bodyStr[start : end+len("</VW_USER_PROFILE>")]
		}
	}

	if profileXML == "" {
		return nil, fmt.Errorf("invalid response from VisionWeb")
	}

	var profile vwUserProfile
	if err := xml.Unmarshal([]byte(profileXML), &profile); err != nil {
		return nil, fmt.Errorf("failed to parse VisionWeb response: %w", err)
	}

	// Check for errors
	if len(profile.Errors.Messages) > 0 {
		msg := profile.Errors.Messages[0]
		return nil, fmt.Errorf("invalid username or password (%s: %s)", msg.ID, msg.Message)
	}

	// Extract labs
	var labs []VWLabInfo
	for _, sup := range profile.CBU.Suppliers {
		for _, loc := range sup.Locations {
			lab := VWLabInfo{
				SupplierName: sup.Name,
				SloID:        loc.ID,
				LocationName: loc.Name,
			}
			if len(loc.Accounts) > 0 {
				lab.BillAccount = loc.Accounts[0].Billing.Number
				lab.ShipAccount = loc.Accounts[0].Shipping.Number
			}
			labs = append(labs, lab)
		}
	}

	return labs, nil
}

func (s *Service) getVWToken(scope string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", scope)

	req, _ := http.NewRequest("POST", vwTokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(vwAPIUser, vwAPIPass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token request failed: %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

func (s *Service) ListIntegrations() ([]map[string]interface{}, error) {
	var companies []intModel.IntegrationCompany
	if err := s.db.Where("active = true").Find(&companies).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(companies))
	for i, c := range companies {
		result[i] = map[string]interface{}{
			"id":   c.IDIntegrationCompany,
			"name": c.Name,
			"code": c.Code,
		}
	}
	return result, nil
}
