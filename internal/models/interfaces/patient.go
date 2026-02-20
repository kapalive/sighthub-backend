// internal/models/interfaces/patient.go
package interfaces

// PatientInterface — чтобы не тянуть пакет patients и не ловить циклы.
// Реализацию даст сервис/репозиторий (например, через GORM).
type PatientInterface interface {
	GetPatientByID(patientID int64) (map[string]interface{}, error)
}
