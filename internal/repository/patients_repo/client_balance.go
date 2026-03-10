// internal/repository/patients_repo/client_balance.go
package patients_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type ClientBalanceRepo struct{ DB *gorm.DB }

func NewClientBalanceRepo(db *gorm.DB) *ClientBalanceRepo {
	return &ClientBalanceRepo{DB: db}
}

// Get возвращает кредитный баланс пациента в локации.
func (r *ClientBalanceRepo) Get(patientID int64, locationID int) (*patients.ClientBalance, error) {
	var row patients.ClientBalance
	err := r.DB.Where("patient_id = ? AND location_id = ?", patientID, locationID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Upsert создаёт или обновляет баланс (по уникальному ключу patient_id+location_id).
func (r *ClientBalanceRepo) Upsert(patientID int64, locationID int, credit float64) error {
	return r.DB.
		Where(patients.ClientBalance{PatientID: patientID, LocationID: locationID}).
		Assign(patients.ClientBalance{Credit: credit}).
		FirstOrCreate(&patients.ClientBalance{}).Error
}

// Add прибавляет к балансу указанную сумму (может быть отрицательной для списания).
func (r *ClientBalanceRepo) Add(patientID int64, locationID int, delta float64) error {
	return r.DB.Exec(
		`INSERT INTO client_balance (patient_id, location_id, credit)
		 VALUES (?, ?, ?)
		 ON CONFLICT (patient_id, location_id) DO UPDATE SET credit = client_balance.credit + ?`,
		patientID, locationID, delta, delta,
	).Error
}

// GetCreditBalance возвращает только числовое значение кредита (удобно для проверок).
func (r *ClientBalanceRepo) GetCreditBalance(patientID int64, locationID int) (float64, error) {
	cb, err := r.Get(patientID, locationID)
	if err != nil || cb == nil {
		return 0, err
	}
	return cb.Credit, nil
}

func (r *ClientBalanceRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
