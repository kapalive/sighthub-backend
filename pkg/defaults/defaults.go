package defaults

// Str returns the pointer value or "" if nil.
func Str(s *string) *string {
	if s != nil {
		return s
	}
	v := ""
	return &v
}

// StrVal returns the string value or "" if nil (non-pointer version).
func StrVal(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// StrDefault returns the pointer value or a pointer to def if nil.
func StrDefault(s *string, def string) *string {
	if s != nil {
		return s
	}
	return &def
}

// Bool returns *bool pointing to the value, or false if nil.
func Bool(b *bool) *bool {
	if b != nil {
		return b
	}
	v := false
	return &v
}

// BoolVal returns the bool value or false if nil.
func BoolVal(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}
