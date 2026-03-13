package license_service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	locModel "sighthub-backend/internal/models/location"
	"sighthub-backend/pkg/email"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

type KMSStoreInput struct {
	Hash    string  `json:"hash"`
	Active  *bool   `json:"active"`
	Message string  `json:"message"`
}

type KMSStoreResult struct {
	Message string `json:"message"`
	Active  bool   `json:"active"`
}

// KMSStore updates the store activation status and optionally sends a notification email.
func (s *Service) KMSStore(input KMSStoreInput) (*KMSStoreResult, error) {
	if input.Hash == "" {
		return nil, fmt.Errorf("hash store is required")
	}

	var store locModel.Store
	if err := s.db.Where("hash = ?", input.Hash).First(&store).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("store not found")
		}
		return nil, err
	}

	var loc locModel.Location
	if err := s.db.Where("store_id = ? AND warehouse_id IS NULL", store.IDStore).First(&loc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("location not found")
		}
		return nil, err
	}

	if input.Active != nil {
		if err := s.db.Model(&loc).Update("store_active", *input.Active).Error; err != nil {
			return nil, err
		}
		loc.StoreActive = input.Active
	}

	msg := input.Message
	if msg == "" {
		msg = "License status changed"
	}

	active := loc.StoreActive != nil && *loc.StoreActive

	// Send notification email if store has an email address (best-effort, ignore errors)
	if store.Email != nil {
		locID := int64(loc.IDLocation)
		_ = email.SendViaDB(s.db, *store.Email, "License update", "default", map[string]interface{}{
			"content": msg,
		}, &locID)
	}

	return &KMSStoreResult{Message: msg, Active: active}, nil
}
