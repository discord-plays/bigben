package utils

func PBool(a bool) *bool { return &a }
func SBool(a *bool) bool { return a != nil && *a }

func PByte(a byte) *byte { return &a }
func SByte(a *byte) byte {
	if a == nil {
		return 0
	}
	return *a
}

func PString(a string) *string { return &a }
func SString(a *string) string {
	if a == nil {
		return ""
	}
	return *a
}
