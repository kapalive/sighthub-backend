// pkg/recall/recall.go
// Аналог utils_recall_sync.py — upsert CL recall в planing_communication
package recall

import (
	"time"

	"gorm.io/gorm"
)

// planingComm — минимальная inline-модель для upsert.
type planingComm struct {
	IDPlaningCommunication int64      `gorm:"column:id_planing_communication;primaryKey;autoIncrement"`
	PatientID              int64      `gorm:"column:patient_id"`
	CommunicationTypeID    int        `gorm:"column:communication_type_id"`
	Reason                 *string    `gorm:"column:reason"`
	Note                   *string    `gorm:"column:note"`
	Date                   time.Time  `gorm:"column:date"`
	LocationID             *int64     `gorm:"column:location_id"`
	SourceTable            *string    `gorm:"column:source_table"`
	SourceID               *int64     `gorm:"column:source_id"`
}

func (planingComm) TableName() string { return "planing_communication" }

// commType — для поиска communication_type_id по имени.
type commType struct {
	CommunicationTypeID int    `gorm:"column:communication_type_id"`
	CommunicationType   string `gorm:"column:communication_type"`
}

func (commType) TableName() string { return "communication_type" }

// UpsertCLRecall создаёт или обновляет запись CL Recall в planing_communication.
// expire_date == zero → удаляет существующую запись.
// Аналог upsert_cl_recall_to_plan из Python.
func UpsertCLRecall(
	db *gorm.DB,
	patientID int64,
	examID int64,
	locationID int64,
	expireDate *time.Time,
	frontDeskNote *string,
) error {
	// Получаем communication_type_id для 'Call'
	var ct commType
	if err := db.Where("communication_type = ?", "Call").First(&ct).Error; err != nil {
		return err
	}

	sourceTable := "eye_exam"
	var existing planingComm
	err := db.Where("source_table = ? AND source_id = ?", sourceTable, examID).
		First(&existing).Error

	found := err == nil

	// Если expire_date не задан — удаляем запись
	if expireDate == nil || expireDate.IsZero() {
		if found {
			return db.Delete(&existing).Error
		}
		return nil
	}

	// Формируем дату recall: expire_date в 09:00 America/New_York (UTC-5)
	// Упрощённо берём UTC, с offset -5h для NY стандартного времени
	recallDT := time.Date(
		expireDate.Year(), expireDate.Month(), expireDate.Day(),
		14, 0, 0, 0, time.UTC, // 09:00 NY = 14:00 UTC (EST)
	)

	reason := "CL Recall"

	if !found {
		st := sourceTable
		sid := examID
		loc := locationID
		rec := planingComm{
			PatientID:           patientID,
			CommunicationTypeID: ct.CommunicationTypeID,
			LocationID:          &loc,
			SourceTable:         &st,
			SourceID:            &sid,
		}
		existing = rec
		existing.Reason = &reason
		existing.Note = frontDeskNote
		existing.Date = recallDT
		return db.Create(&existing).Error
	}

	// Обновляем
	existing.Reason = &reason
	existing.Note = frontDeskNote
	existing.Date = recallDT
	return db.Save(&existing).Error
}
