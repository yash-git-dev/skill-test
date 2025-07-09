package models

import "time"

// Student represents the student data structure from the Node.js API
type Student struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	Email              string  `json:"email"`
	SystemAccess       bool    `json:"systemAccess"`
	Phone              *string `json:"phone"`
	Gender             *string `json:"gender"`
	DOB                *string `json:"dob"`
	Class              *string `json:"class"`
	Section            *string `json:"section"`
	Roll               *int    `json:"roll"`
	FatherName         *string `json:"fatherName"`
	FatherPhone        *string `json:"fatherPhone"`
	MotherName         *string `json:"motherName"`
	MotherPhone        *string `json:"motherPhone"`
	GuardianName       *string `json:"guardianName"`
	GuardianPhone      *string `json:"guardianPhone"`
	RelationOfGuardian *string `json:"relationOfGuardian"`
	CurrentAddress     *string `json:"currentAddress"`
	PermanentAddress   *string `json:"permanentAddress"`
	AdmissionDate      *string `json:"admissionDate"`
	ReporterName       *string `json:"reporterName"`
}

// APIResponse represents the standardized API response from Node.js backend
type APIResponse struct {
	Success bool    `json:"success"`
	Data    Student `json:"data"`
	Message string  `json:"message"`
}

// ErrorResponse represents error response structure
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ReportMetadata contains metadata for PDF generation
type ReportMetadata struct {
	GeneratedAt time.Time `json:"generated_at"`
	GeneratedBy string    `json:"generated_by"`
	ReportID    string    `json:"report_id"`
}

// StudentListResponse represents the response for listing students
type StudentListResponse struct {
	Success bool              `json:"success"`
	Data    []StudentListItem `json:"data"`
	Message string            `json:"message"`
}

// StudentListItem represents a student in the list view (simplified)
type StudentListItem struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	SystemAccess bool    `json:"systemAccess"`
	Class        *string `json:"class"`
	Section      *string `json:"section"`
	Roll         *int    `json:"roll"`
}

// FormatName safely returns the student name or "N/A" if empty
func (s *Student) FormatName() string {
	if s.Name != "" {
		return s.Name
	}
	return "N/A"
}

// FormatEmail safely returns the student email or "N/A" if empty
func (s *Student) FormatEmail() string {
	if s.Email != "" {
		return s.Email
	}
	return "N/A"
}

// SafeString returns the value of a string pointer or defaultValue if nil
func SafeString(ptr *string, defaultValue string) string {
	if ptr != nil && *ptr != "" {
		return *ptr
	}
	return defaultValue
}

// SafeInt returns the value of an int pointer or defaultValue if nil
func SafeInt(ptr *int, defaultValue int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}
