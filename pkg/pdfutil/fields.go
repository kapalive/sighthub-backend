// pkg/pdfutil/fields.go
// Чтение AcroForm-полей PDF — аналог get_pdf_fields из utils_form.py
package pdfutil

import (
	"bytes"
	"os"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/form"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// FieldInfo описывает одно AcroForm-поле PDF.
type FieldInfo struct {
	ID    string // имя/идентификатор поля
	Type  string // "text", "date", "checkbox", "radio", "combobox", "listbox"
	Value string // текущее значение (может быть пустым)
}

// GetFields возвращает список всех AcroForm-полей из PDF-файла.
// Аналог get_pdf_fields из Python (PyPDF2).
func GetFields(pdfPath string) ([]FieldInfo, error) {
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, err
	}
	return GetFieldsBytes(data)
}

// GetFieldsBytes читает AcroForm-поля из PDF-байтов (без создания файла).
func GetFieldsBytes(data []byte) ([]FieldInfo, error) {
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	fields, err := pdfcpuapi.FormFields(bytes.NewReader(data), conf)
	if err != nil {
		return nil, err
	}
	return convertFields(fields), nil
}

// GetFieldMap возвращает map[fieldID]currentValue для удобного поиска.
func GetFieldMap(pdfPath string) (map[string]string, error) {
	fields, err := GetFields(pdfPath)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(fields))
	for _, f := range fields {
		m[f.ID] = f.Value
	}
	return m, nil
}

// convertFields переводит []form.Field в []FieldInfo.
func convertFields(fields []form.Field) []FieldInfo {
	out := make([]FieldInfo, 0, len(fields))
	for _, f := range fields {
		out = append(out, FieldInfo{
			ID:    f.ID,
			Type:  fieldTypeName(f.Typ),
			Value: f.V,
		})
	}
	return out
}

func fieldTypeName(t form.FieldType) string {
	switch t {
	case form.FTText:
		return "text"
	case form.FTDate:
		return "date"
	case form.FTCheckBox:
		return "checkbox"
	case form.FTComboBox:
		return "combobox"
	case form.FTListBox:
		return "listbox"
	case form.FTRadioButtonGroup:
		return "radio"
	default:
		return "unknown"
	}
}

// openForRead открывает файл для чтения (используется внутри).
func openForRead(path string) (*os.File, error) {
	return os.Open(path)
}
