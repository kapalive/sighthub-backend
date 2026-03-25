package zeiss_service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInvalidState = fmt.Errorf("invalid or expired state parameter")

// ─── Config ──────────────────────────────────────────────────────────────────

type zeissConfig struct {
	ClientID    string
	AuthURL     string
	TokenURL    string
	RedirectURI string
	APIBase     string
	OrderSubKey string
	PCatSubKey  string
}

func loadConfig() zeissConfig {
	return zeissConfig{
		ClientID:    envOrDefault("ZEISS_CLIENT_ID", "a346b369-8dd8-4fe4-a8bc-755a5192ead6"),
		AuthURL:     envOrDefault("ZEISS_AUTH_URL", "https://id-ip-stage.zeiss.com/B2C_1A_ZeissIdNormalSignIn/oauth2/v2.0/authorize"),
		TokenURL:    envOrDefault("ZEISS_TOKEN_URL", "https://id-ip-stage.zeiss.com/B2C_1A_ZeissIdNormalSignIn/oauth2/v2.0/token"),
		RedirectURI: os.Getenv("ZEISS_REDIRECT_URI"),
		APIBase:     envOrDefault("ZEISS_API_BASE", "https://api-stage.vision.zeiss.com"),
		OrderSubKey: envOrDefault("ZEISS_ORDER_SUB_KEY", "8d1181e5097f40e981fc802f65f03814"),
		PCatSubKey:  envOrDefault("ZEISS_PCAT_SUB_KEY", "2378ceefecea4a52b6b084cf2a52c980"),
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// ─── Model ───────────────────────────────────────────────────────────────────

type ZeissToken struct {
	ID             int64     `gorm:"column:id_zeiss_token;primaryKey"`
	EmployeeID     int64     `gorm:"column:employee_id;uniqueIndex"`
	AccessToken    string    `gorm:"column:access_token"`
	RefreshToken   *string   `gorm:"column:refresh_token"`
	CustomerNumber *string   `gorm:"column:customer_number"`
	ExpiresAt      time.Time `gorm:"column:expires_at"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (ZeissToken) TableName() string { return "zeiss_token" }

// ─── Auth Service ────────────────────────────────────────────────────────────

type AuthService struct {
	db    *gorm.DB
	redis *redis.Client
	cfg   zeissConfig
}

func NewAuthService(db *gorm.DB, rdb *redis.Client) *AuthService {
	return &AuthService{
		db:    db,
		redis: rdb,
		cfg:   loadConfig(),
	}
}

const pkceTTL = 10 * time.Minute

// pkceState is the JSON structure stored in Redis for the PKCE flow.
type pkceState struct {
	CodeVerifier string `json:"code_verifier"`
	EmployeeID   int64  `json:"employee_id"`
	RedirectURI  string `json:"redirect_uri"`
}

// ─── GenerateAuthURL ─────────────────────────────────────────────────────────

func (s *AuthService) GenerateAuthURL(employeeID int64, origin string) (map[string]string, error) {
	ctx := context.Background()

	// Generate PKCE code verifier (32 random bytes → base64url)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("zeiss auth: generate code_verifier: %w", err)
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Code challenge = base64url(sha256(code_verifier))
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// State = 16 random bytes → base64url
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("zeiss auth: generate state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	// Store PKCE data in Redis
	redirectURI := origin + "/zeiss/callback"
	data, err := json.Marshal(pkceState{
		CodeVerifier: codeVerifier,
		EmployeeID:   employeeID,
		RedirectURI:  redirectURI,
	})
	if err != nil {
		return nil, fmt.Errorf("zeiss auth: marshal pkce state: %w", err)
	}

	redisKey := fmt.Sprintf("zeiss_pkce:%s", state)
	if err := s.redis.Set(ctx, redisKey, data, pkceTTL).Err(); err != nil {
		return nil, fmt.Errorf("zeiss auth: store pkce state: %w", err)
	}

	// Build auth URL
	scope := fmt.Sprintf("openid %s offline_access", s.cfg.ClientID)
	params := url.Values{
		"p":                     {"B2C_1A_ZeissIdNormalSignIn"},
		"response_type":         {"code"},
		"client_id":             {s.cfg.ClientID},
		"scope":                 {scope},
		"redirect_uri":          {origin + "/zeiss/callback"},
		"response_mode":         {"query"},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}

	authURL := fmt.Sprintf("%s?%s", s.cfg.AuthURL, params.Encode())

	return map[string]string{
		"auth_url": authURL,
		"state":    state,
	}, nil
}

// ─── ExchangeCode ────────────────────────────────────────────────────────────

func (s *AuthService) ExchangeCode(state string, code string) error {
	ctx := context.Background()

	// Retrieve PKCE state from Redis
	redisKey := fmt.Sprintf("zeiss_pkce:%s", state)
	raw, err := s.redis.Get(ctx, redisKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrInvalidState
		}
		return fmt.Errorf("zeiss auth: redis get: %w", err)
	}

	var ps pkceState
	if err := json.Unmarshal(raw, &ps); err != nil {
		return fmt.Errorf("zeiss auth: unmarshal pkce state: %w", err)
	}

	// Exchange code for tokens
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {ps.RedirectURI},
		"code_verifier": {ps.CodeVerifier},
	}

	tokenResp, err := s.postTokenRequest(ctx, form)
	if err != nil {
		return err
	}

	// Calculate expiry
	expiresInSec, _ := tokenResp.ExpiresIn.Int64()
	if expiresInSec <= 0 {
		expiresInSec = 3600
	}
	expiresAt := time.Now().UTC().Add(time.Duration(expiresInSec) * time.Second)

	// Extract customer_number from JWT payload
	custNum := extractCustomerNumber(tokenResp.AccessToken)

	// Upsert token
	now := time.Now().UTC()
	tok := ZeissToken{
		EmployeeID:     ps.EmployeeID,
		AccessToken:    tokenResp.AccessToken,
		CustomerNumber: custNum,
		ExpiresAt:      expiresAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if tokenResp.RefreshToken != "" {
		tok.RefreshToken = &tokenResp.RefreshToken
	}

	err = s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "employee_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"access_token", "refresh_token", "customer_number", "expires_at", "updated_at",
		}),
	}).Create(&tok).Error
	if err != nil {
		return fmt.Errorf("zeiss auth: upsert token: %w", err)
	}

	// Clean up Redis
	s.redis.Del(ctx, redisKey)

	return nil
}

// ─── GetToken ────────────────────────────────────────────────────────────────

func (s *AuthService) GetToken(employeeID int64) (string, error) {
	ctx := context.Background()

	var tok ZeissToken
	if err := s.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		First(&tok).Error; err != nil {
		return "", fmt.Errorf("zeiss auth required")
	}

	// If expired, try to refresh
	if time.Now().UTC().After(tok.ExpiresAt) {
		if tok.RefreshToken == nil || *tok.RefreshToken == "" {
			return "", fmt.Errorf("zeiss auth required")
		}
		if err := s.refreshToken(ctx, employeeID, *tok.RefreshToken); err != nil {
			return "", fmt.Errorf("zeiss auth required")
		}
		// Reload after refresh
		if err := s.db.WithContext(ctx).
			Where("employee_id = ?", employeeID).
			First(&tok).Error; err != nil {
			return "", fmt.Errorf("zeiss auth required")
		}
	}

	return tok.AccessToken, nil
}

// ─── RefreshToken ────────────────────────────────────────────────────────────

func (s *AuthService) RefreshToken(employeeID int64, refreshToken string) error {
	return s.refreshToken(context.Background(), employeeID, refreshToken)
}

func (s *AuthService) refreshToken(ctx context.Context, employeeID int64, refreshToken string) error {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {s.cfg.ClientID},
	}

	tokenResp, err := s.postTokenRequest(ctx, form)
	if err != nil {
		return err
	}

	expiresInSec, _ := tokenResp.ExpiresIn.Int64()
	if expiresInSec <= 0 {
		expiresInSec = 3600
	}
	expiresAt := time.Now().UTC().Add(time.Duration(expiresInSec) * time.Second)
	now := time.Now().UTC()

	updates := map[string]interface{}{
		"access_token": tokenResp.AccessToken,
		"expires_at":   expiresAt,
		"updated_at":   now,
	}
	if tokenResp.RefreshToken != "" {
		updates["refresh_token"] = tokenResp.RefreshToken
	}

	return s.db.WithContext(ctx).
		Model(&ZeissToken{}).
		Where("employee_id = ?", employeeID).
		Updates(updates).Error
}

// ─── IsAuthenticated / GetAuthStatus ─────────────────────────────────────────

func (s *AuthService) IsAuthenticated(employeeID int64) bool {
	var count int64
	s.db.WithContext(context.Background()).
		Model(&ZeissToken{}).
		Where("employee_id = ?", employeeID).
		Count(&count)
	return count > 0
}

type AuthStatus struct {
	Authenticated  bool    `json:"authenticated"`
	CustomerNumber *string `json:"customer_number"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
}

func (s *AuthService) GetAuthStatus(employeeID int64) AuthStatus {
	var tok ZeissToken
	err := s.db.Where("employee_id = ?", employeeID).First(&tok).Error
	if err != nil {
		return AuthStatus{Authenticated: false}
	}
	exp := tok.ExpiresAt.Format(time.RFC3339)
	return AuthStatus{
		Authenticated:  true,
		CustomerNumber: tok.CustomerNumber,
		ExpiresAt:      &exp,
	}
}

func (s *AuthService) Logout(employeeID int64) {
	s.db.Where("employee_id = ?", employeeID).Delete(&ZeissToken{})
}

// extractCustomerNumber parses JWT payload to get zeissSecondaryReference
func extractCustomerNumber(accessToken string) *string {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil
	}
	// Decode payload (base64url, may need padding)
	payload := parts[1]
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil
	}
	// Parse ZeissIdOrganisation (comes as escaped JSON string in the claim)
	var claims map[string]interface{}
	if json.Unmarshal(decoded, &claims) != nil {
		return nil
	}
	orgStr, ok := claims["ZeissIdOrganisation"].(string)
	if !ok {
		return nil
	}
	var org struct {
		ZeissSecondaryReference string `json:"zeissSecondaryReference"`
	}
	if json.Unmarshal([]byte(orgStr), &org) != nil || org.ZeissSecondaryReference == "" {
		return nil
	}
	return &org.ZeissSecondaryReference
}

// ─── HTTP helpers ────────────────────────────────────────────────────────────

type authTokenResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresIn    json.Number     `json:"expires_in"`
}

func (s *AuthService) APIBase() string    { return s.cfg.APIBase }
func (s *AuthService) PCatSubKey() string  { return s.cfg.PCatSubKey }
func (s *AuthService) OrderSubKey() string { return s.cfg.OrderSubKey }

func (s *AuthService) postTokenRequest(ctx context.Context, form url.Values) (*authTokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("zeiss auth: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zeiss auth: token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("zeiss auth: token request failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var data authTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("zeiss auth: decode token response: %w", err)
	}

	return &data, nil
}
