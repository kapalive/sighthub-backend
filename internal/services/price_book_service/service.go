package price_book_service

import (
	"fmt"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ─── shared helpers ───────────────────────────────────────────────────────────

func strOrZero(v *float64) string {
	if v == nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", *v)
}

func strOrNil(v *float64) *string {
	if v == nil {
		return nil
	}
	s := fmt.Sprintf("%.2f", *v)
	return &s
}

func parseDec(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case string:
		var f float64
		_, err := fmt.Sscanf(t, "%f", &f)
		return f, err == nil
	}
	return 0, false
}
