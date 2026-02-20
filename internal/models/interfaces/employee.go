// internal/models/interfaces/employee.go
package interfaces

// EmployeeInterface - интерфейс для работы с сотрудниками
type EmployeeInterface interface {
	GetEmployeeByID(employeeID int64) (map[string]interface{}, error)
}
