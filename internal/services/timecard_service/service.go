package timecard_service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/models/audit"
	"sighthub-backend/internal/models/employees"
	pkgAuth "sighthub-backend/pkg/auth"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidCreds    = errors.New("invalid username or password")
	ErrInactive        = errors.New("account is inactive")
	ErrHistoryNotFound = errors.New("history entry not found")
)

type Service struct {
	db  *gorm.DB
	rdb *redis.Client
	cfg *config.Config
}

func New(db *gorm.DB, rdb *redis.Client, cfg *config.Config) *Service {
	return &Service{db: db, rdb: rdb, cfg: cfg}
}

func (s *Service) getUser(username string) (*employees.EmployeeTimecardLogin, error) {
	var user employees.EmployeeTimecardLogin
	err := s.db.Where("lower(username) = ?", strings.ToLower(username)).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

// ─── Session Status ───────────────────────────────────────────────────────────

type SessionStatusResult struct {
	Message     string `json:"message"`
	LastAction  string `json:"last_action"`
	CurrentTime string `json:"current_time"`
	Action      string `json:"action"`
}

func (s *Service) GetSessionStatus(username string) (*SessionStatusResult, error) {
	user, err := s.getUser(username)
	if err != nil {
		return nil, err
	}

	var last audit.EmployeeTimecardHistory
	err = s.db.
		Where("employee_timecard_login_id = ?", user.IDEmployeeTimecardLogin).
		Order("timestamp desc").
		First(&last).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrHistoryNotFound
	}
	if err != nil {
		return nil, err
	}

	lastAction := last.Timestamp.Format("2006-01-02 15:04:05")
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return &SessionStatusResult{
		Message:     fmt.Sprintf("Successfully %s at %s", last.ActionType, lastAction),
		LastAction:  lastAction,
		CurrentTime: currentTime,
		Action:      last.ActionType,
	}, nil
}

// ─── Check In/Out ─────────────────────────────────────────────────────────────

type CheckInOutResult struct {
	Message string `json:"message"`
	Time    string `json:"time"`
	Action  string `json:"action"`
}

func (s *Service) CheckInOut(username, password string) (*CheckInOutResult, error) {
	user, err := s.getUser(username)
	if err != nil {
		return nil, ErrInvalidCreds
	}
	if !user.Active {
		return nil, ErrInactive
	}
	if !user.CheckPassword(password) {
		return nil, ErrInvalidCreds
	}

	var last audit.EmployeeTimecardHistory
	err = s.db.
		Where("employee_timecard_login_id = ?", user.IDEmployeeTimecardLogin).
		Order("timestamp desc").
		First(&last).Error

	actionType := "checkin"
	if err == nil && last.ActionType == "checkin" {
		actionType = "checkout"
	}

	entry := audit.EmployeeTimecardHistory{
		EmployeeTimecardLoginID: user.IDEmployeeTimecardLogin,
		ActionType:              actionType,
		Timestamp:               time.Now(),
	}
	if err := s.db.Create(&entry).Error; err != nil {
		return nil, err
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return &CheckInOutResult{
		Message: fmt.Sprintf("Successfully %s at %s", actionType, currentTime),
		Time:    currentTime,
		Action:  actionType,
	}, nil
}

// ─── Login ────────────────────────────────────────────────────────────────────

func (s *Service) Login(username, password string) (string, error) {
	user, err := s.getUser(strings.ToLower(username))
	if err != nil {
		return "", ErrInvalidCreds
	}
	if !user.CheckPassword(password) {
		return "", ErrInvalidCreds
	}
	if !user.Active {
		return "", ErrInactive
	}

	token, _, err := pkgAuth.GenerateAccessToken(user.Username, s.cfg.JWTSecretKey)
	return token, err
}

// ─── Info ─────────────────────────────────────────────────────────────────────

type PeriodRecord struct {
	ID       int     `json:"id"`
	Checkin  string  `json:"checkin"`
	Checkout string  `json:"checkout"`
	Summary  string  `json:"summary"`
	Note     *string `json:"note"`
}

type InfoResult struct {
	TotalTime string         `json:"total_time"`
	Periods   []PeriodRecord `json:"periods"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
}

func (s *Service) GetInfo(username string, startDate, endDate time.Time) (*InfoResult, error) {
	user, err := s.getUser(username)
	if err != nil {
		return nil, err
	}

	var history []audit.EmployeeTimecardHistory
	if err := s.db.
		Where("employee_timecard_login_id = ? AND timestamp >= ? AND timestamp <= ?",
			user.IDEmployeeTimecardLogin, startDate, endDate).
		Order("timestamp asc").
		Find(&history).Error; err != nil {
		return nil, err
	}

	type periodEntry struct {
		checkinTime time.Time
		record      PeriodRecord
	}

	var periods []periodEntry
	var totalSeconds float64
	var lastCheckin *audit.EmployeeTimecardHistory

	for i := range history {
		entry := &history[i]
		if entry.ActionType == "checkin" {
			lastCheckin = entry
		} else if entry.ActionType == "checkout" && lastCheckin != nil {
			diff := entry.Timestamp.Sub(lastCheckin.Timestamp)
			totalSeconds += diff.Seconds()
			sec := int(diff.Seconds())
			periods = append(periods, periodEntry{
				checkinTime: lastCheckin.Timestamp,
				record: PeriodRecord{
					ID:       lastCheckin.IDEmployeeTimecardHistory,
					Checkin:  lastCheckin.Timestamp.Format("15:04:05 01/02/2006"),
					Checkout: entry.Timestamp.Format("15:04:05 01/02/2006"),
					Summary:  fmt.Sprintf("%d:%02d", sec/3600, (sec%3600)/60),
					Note:     lastCheckin.Note,
				},
			})
			lastCheckin = nil
		}
	}

	sort.Slice(periods, func(i, j int) bool {
		return periods[i].checkinTime.After(periods[j].checkinTime)
	})

	result := make([]PeriodRecord, len(periods))
	for i, p := range periods {
		result[i] = p.record
	}

	total := int(totalSeconds)
	return &InfoResult{
		TotalTime: fmt.Sprintf("%d:%02d", total/3600, (total%3600)/60),
		Periods:   result,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}

// ─── Change Password ──────────────────────────────────────────────────────────

func (s *Service) ChangePassword(username, oldPassword, newPassword string) error {
	user, err := s.getUser(username)
	if err != nil {
		return ErrInvalidCreds
	}
	if !user.CheckPassword(oldPassword) {
		return ErrInvalidCreds
	}
	if !user.Active {
		return ErrInactive
	}
	if err := user.SetPassword(newPassword); err != nil {
		return err
	}
	return s.db.Save(user).Error
}

// ─── Update Note ──────────────────────────────────────────────────────────────

func (s *Service) UpdateHistoryNote(username string, historyID int, note *string) error {
	user, err := s.getUser(username)
	if err != nil {
		return err
	}

	var entry audit.EmployeeTimecardHistory
	err = s.db.First(&entry, historyID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || entry.EmployeeTimecardLoginID != user.IDEmployeeTimecardLogin {
		return ErrHistoryNotFound
	}
	if err != nil {
		return err
	}

	entry.Note = note
	return s.db.Save(&entry).Error
}
