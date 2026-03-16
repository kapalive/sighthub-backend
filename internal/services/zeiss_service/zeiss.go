package zeiss_service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	intModel "sighthub-backend/internal/models/integrations"
)

var (
	ErrCredentialNotFound = errors.New("integration credential not found")
	ErrCredentialInactive = errors.New("integration credential is inactive")
	ErrMissingCode        = errors.New("missing code or state")
	ErrInvalidState       = errors.New("invalid or expired state")
	ErrTokenExchange      = errors.New("token exchange failed")
	ErrNotAuthenticated   = errors.New("not authenticated with ZEISS")
	ErrNoRefreshToken     = errors.New("no refresh token available")
)

const oauthStateTTL = 10 * time.Minute

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ─── GetAuthURL ──────────────────────────────────────────────────────────────

type AuthURLResponse struct {
	AuthURL string `json:"authUrl"`
}

func (s *Service) GetAuthURL(ctx context.Context, locationID int64) (*AuthURLResponse, error) {
	cred, zc, err := s.loadZeissCredential(ctx, locationID)
	if err != nil {
		return nil, err
	}

	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, err
	}
	codeChallenge := pkceCodeChallenge(codeVerifier)

	state := fmt.Sprintf("%s_%d", zc.OAuthState, cred.IDIntegrationCredential)

	// persist PKCE state
	oauthState := intModel.IntegrationOAuthState{
		IntegrationCredentialID: cred.IDIntegrationCredential,
		State:                   state,
		CodeVerifier:            codeVerifier,
		CreatedAt:               time.Now().UTC(),
		ExpiresAt:               time.Now().UTC().Add(oauthStateTTL),
	}
	if err := s.db.WithContext(ctx).Create(&oauthState).Error; err != nil {
		return nil, err
	}

	scope := fmt.Sprintf("openid %s offline_access", cred.ClientID)

	redirectURI := fmt.Sprintf("%s/api/integrations/zeiss/oauth2redirect", strings.TrimRight(zc.BaseURL, "/"))

	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {cred.ClientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {scope},
		"response_mode":         {"query"},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
		"p":                     {zc.Policy},
	}

	authURL := fmt.Sprintf("%s?%s", zc.AuthorizeURL, params.Encode())
	return &AuthURLResponse{AuthURL: authURL}, nil
}

// ─── HandleCallback ──────────────────────────────────────────────────────────

func (s *Service) HandleCallback(ctx context.Context, code, stateParam string) (redirectURL string, err error) {
	if code == "" || stateParam == "" {
		return "", ErrMissingCode
	}

	oauthState, err := s.consumeOAuthState(ctx, stateParam)
	if err != nil {
		return "", err
	}

	cred, zc, err := s.loadZeissCredentialByID(ctx, oauthState.IntegrationCredentialID)
	if err != nil {
		return "", err
	}

	redirectURI := fmt.Sprintf("%s/api/integrations/zeiss/oauth2redirect", strings.TrimRight(zc.BaseURL, "/"))

	tokenData, err := s.doTokenExchange(ctx, cred, zc, code, redirectURI, oauthState.CodeVerifier)
	if err != nil {
		frontPath := strings.TrimRight(zc.BaseURL, "/") + zc.FrontRedirectPath
		return frontPath + "?zeiss_error=token_exchange_failed", nil
	}

	if err := s.storeTokens(ctx, cred.IDIntegrationCredential, tokenData); err != nil {
		return "", err
	}

	frontPath := strings.TrimRight(zc.BaseURL, "/") + zc.FrontRedirectPath
	return frontPath, nil
}

// ─── Exchange (hybrid flow) ──────────────────────────────────────────────────

type ExchangeRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func (s *Service) Exchange(ctx context.Context, req ExchangeRequest) error {
	if req.Code == "" || req.State == "" {
		return ErrMissingCode
	}

	oauthState, err := s.consumeOAuthState(ctx, req.State)
	if err != nil {
		return err
	}

	cred, zc, err := s.loadZeissCredentialByID(ctx, oauthState.IntegrationCredentialID)
	if err != nil {
		return err
	}

	redirectURI := fmt.Sprintf("%s%s", strings.TrimRight(zc.BaseURL, "/"), zc.FrontCallbackPath)

	tokenData, err := s.doTokenExchange(ctx, cred, zc, req.Code, redirectURI, oauthState.CodeVerifier)
	if err != nil {
		return err
	}

	return s.storeTokens(ctx, cred.IDIntegrationCredential, tokenData)
}

// ─── GetToken ────────────────────────────────────────────────────────────────

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
}

func (s *Service) GetToken(ctx context.Context, locationID int64) (*TokenResponse, error) {
	cred, _, err := s.loadZeissCredential(ctx, locationID)
	if err != nil {
		return nil, err
	}

	var tok intModel.IntegrationToken
	if err := s.db.WithContext(ctx).
		Where("integration_credential_id = ?", cred.IDIntegrationCredential).
		First(&tok).Error; err != nil {
		return nil, ErrNotAuthenticated
	}

	resp := &TokenResponse{
		AccessToken: tok.AccessToken,
	}
	if tok.RefreshToken != nil {
		resp.RefreshToken = *tok.RefreshToken
	}
	if tok.ExpiresIn != nil {
		resp.ExpiresIn = *tok.ExpiresIn
	}
	return resp, nil
}

// ─── RefreshToken ────────────────────────────────────────────────────────────

func (s *Service) RefreshToken(ctx context.Context, locationID int64) (*TokenResponse, error) {
	cred, zc, err := s.loadZeissCredential(ctx, locationID)
	if err != nil {
		return nil, err
	}

	var tok intModel.IntegrationToken
	if err := s.db.WithContext(ctx).
		Where("integration_credential_id = ?", cred.IDIntegrationCredential).
		First(&tok).Error; err != nil {
		return nil, ErrNotAuthenticated
	}
	if tok.RefreshToken == nil || *tok.RefreshToken == "" {
		return nil, ErrNoRefreshToken
	}

	scope := fmt.Sprintf("openid %s offline_access", cred.ClientID)

	tokenURL := zc.TokenURL
	if zc.Policy != "" {
		tokenURL = fmt.Sprintf("%s?p=%s", tokenURL, zc.Policy)
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {cred.ClientID},
		"scope":         {scope},
		"refresh_token": {*tok.RefreshToken},
	}

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("zeiss refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("zeiss refresh failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var data tokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if err := s.storeTokens(ctx, cred.IDIntegrationCredential, &data); err != nil {
		return nil, err
	}

	result := &TokenResponse{
		AccessToken: data.AccessToken,
	}
	if data.RefreshToken != "" {
		result.RefreshToken = data.RefreshToken
	} else if tok.RefreshToken != nil {
		result.RefreshToken = *tok.RefreshToken
	}
	if data.ExpiresIn != nil {
		result.ExpiresIn = *data.ExpiresIn
	}
	return result, nil
}

// ─── Logout ──────────────────────────────────────────────────────────────────

func (s *Service) Logout(ctx context.Context, locationID int64) error {
	cred, _, err := s.loadZeissCredential(ctx, locationID)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).
		Where("integration_credential_id = ?", cred.IDIntegrationCredential).
		Delete(&intModel.IntegrationToken{}).Error
}

// ─── private helpers ─────────────────────────────────────────────────────────

func (s *Service) loadZeissCredential(ctx context.Context, locationID int64) (*intModel.IntegrationCredential, *intModel.ZeissConfig, error) {
	var cred intModel.IntegrationCredential
	err := s.db.WithContext(ctx).
		Where("provider = ? AND (location_id = ? OR location_id IS NULL)", "zeiss", locationID).
		Order("location_id DESC NULLS LAST").
		First(&cred).Error
	if err != nil {
		return nil, nil, ErrCredentialNotFound
	}
	if !cred.Active {
		return nil, nil, ErrCredentialInactive
	}
	zc, err := cred.ParseZeissConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("invalid zeiss config: %w", err)
	}
	return &cred, zc, nil
}

func (s *Service) loadZeissCredentialByID(ctx context.Context, id int64) (*intModel.IntegrationCredential, *intModel.ZeissConfig, error) {
	var cred intModel.IntegrationCredential
	if err := s.db.WithContext(ctx).First(&cred, id).Error; err != nil {
		return nil, nil, ErrCredentialNotFound
	}
	if !cred.Active {
		return nil, nil, ErrCredentialInactive
	}
	zc, err := cred.ParseZeissConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("invalid zeiss config: %w", err)
	}
	return &cred, zc, nil
}

func (s *Service) consumeOAuthState(ctx context.Context, stateParam string) (*intModel.IntegrationOAuthState, error) {
	var oauthState intModel.IntegrationOAuthState
	err := s.db.WithContext(ctx).
		Where("state = ? AND expires_at > ?", stateParam, time.Now().UTC()).
		First(&oauthState).Error
	if err != nil {
		return nil, ErrInvalidState
	}
	// delete consumed state
	s.db.WithContext(ctx).Delete(&oauthState)
	return &oauthState, nil
}

type tokenExchangeResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    *int64 `json:"expires_in"`
}

func (s *Service) doTokenExchange(ctx context.Context, cred *intModel.IntegrationCredential, zc *intModel.ZeissConfig, code, redirectURI, codeVerifier string) (*tokenExchangeResponse, error) {
	tokenURL := zc.TokenURL
	if zc.Policy != "" {
		tokenURL = fmt.Sprintf("%s?p=%s", tokenURL, zc.Policy)
	}

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zeiss token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status=%d body=%s", ErrTokenExchange, resp.StatusCode, string(body))
	}

	var data tokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *Service) storeTokens(ctx context.Context, credentialID int64, data *tokenExchangeResponse) error {
	now := time.Now().UTC()

	var existing intModel.IntegrationToken
	err := s.db.WithContext(ctx).
		Where("integration_credential_id = ?", credentialID).
		First(&existing).Error

	if err == nil {
		// update existing
		existing.AccessToken = data.AccessToken
		if data.RefreshToken != "" {
			existing.RefreshToken = &data.RefreshToken
		}
		existing.ExpiresIn = data.ExpiresIn
		existing.ObtainedAt = now
		existing.UpdatedAt = now
		return s.db.WithContext(ctx).Save(&existing).Error
	}

	// create new
	tok := intModel.IntegrationToken{
		IntegrationCredentialID: credentialID,
		AccessToken:             data.AccessToken,
		ExpiresIn:               data.ExpiresIn,
		ObtainedAt:              now,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if data.RefreshToken != "" {
		tok.RefreshToken = &data.RefreshToken
	}
	return s.db.WithContext(ctx).Create(&tok).Error
}

// ─── PKCE helpers ────────────────────────────────────────────────────────────

func generateCodeVerifier() (string, error) {
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func pkceCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
