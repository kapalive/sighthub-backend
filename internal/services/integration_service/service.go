package integration_service

import (
	"encoding/json"
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
	vwAPIUser   = "qademo"
	vwAPIPass   = "xNH3L7K1wl9lezUZQwdA"
	vwTokenURL  = "https://regauth.visionweb.com/connect/token"
	vwStatusURL = "https://regapi.visionweb.com/order/OrderTracking/GetTrackingUpdates"
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

	switch code {
	case "vw":
		ok, err := s.checkVisionWebAuth(username, password)
		connected = ok
		if err != nil {
			checkErr = err.Error()
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

	return result, nil
}

func (s *Service) checkVisionWebAuth(username, password string) (bool, error) {
	token, err := s.getVWToken("VW.OP.Order.WebApi")
	if err != nil {
		return false, fmt.Errorf("VisionWeb API auth failed: %w", err)
	}

	body := `{"OrderIds": [], "StartDate": "2026-01-01", "EndDate": "2026-01-02"}`
	req, _ := http.NewRequest("POST", vwStatusURL, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("username", username)
	req.Header.Set("password", password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("VisionWeb connection failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return false, fmt.Errorf("invalid username or password")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err == nil {
		if errMsg, ok := result["Error"]; ok && errMsg != nil {
			errStr := fmt.Sprintf("%v", errMsg)
			if strings.Contains(strings.ToLower(errStr), "auth") || strings.Contains(strings.ToLower(errStr), "credential") {
				return false, fmt.Errorf("invalid username or password")
			}
		}
	}

	return true, nil
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
