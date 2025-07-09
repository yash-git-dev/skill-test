package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"student-report-service/internal/service"

	"github.com/gorilla/mux"
)

// StudentPDFHandler handles HTTP requests for report generation
type StudentPDFHandler struct {
	pdfService *service.PDFReportService
}

// NewStudentPDFHandler creates a new report handler
func NewStudentPDFHandler(pdfService *service.PDFReportService) *StudentPDFHandler {
	return &StudentPDFHandler{
		pdfService: pdfService,
	}
}

// CreateStudentPDF handles POST /api/v1/reports/student/{id}
func (h *StudentPDFHandler) CreateStudentPDF(w http.ResponseWriter, r *http.Request) {
	// Extract student ID from URL
	vars := mux.Vars(r)
	studentIDStr, exists := vars["id"]
	if !exists {
		h.writeErrorResponse(w, http.StatusBadRequest, "Student ID is required", nil)
		return
	}

	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil || studentID <= 0 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid student ID format", err)
		return
	}

	// Get generated_by from query params or default to "API"
	generatedBy := r.URL.Query().Get("generated_by")
	if generatedBy == "" {
		generatedBy = "API"
	}

	// Generate the report
	result, err := h.pdfService.CreateStudentPDF(studentID, generatedBy)
	if err != nil {
		statusCode := http.StatusInternalServerError

		// Check if it's a client error (student not found, etc.)
		if isClientError(err) {
			statusCode = http.StatusNotFound
		}

		h.writeErrorResponse(w, statusCode, "Failed to generate report", err)
		return
	}

	// Return success response
	h.writeSuccessResponse(w, http.StatusCreated, "Report generated successfully", result)
}

// HealthCheck handles GET /health
func (h *StudentPDFHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := h.pdfService.HealthCheck()

	statusCode := http.StatusOK
	if !status.Healthy {
		statusCode = http.StatusServiceUnavailable
	}

	h.writeResponse(w, statusCode, status)
}

// CleanupReports handles POST /api/v1/reports/cleanup
func (h *StudentPDFHandler) CleanupReports(w http.ResponseWriter, r *http.Request) {
	err := h.pdfService.CleanupOldReports()
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to cleanup reports", err)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"message":   "Old reports cleaned up successfully",
		"timestamp": time.Now(),
	}

	h.writeResponse(w, http.StatusOK, response)
}

// GetStudents handles GET /api/v1/students
func (h *StudentPDFHandler) GetStudents(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters for filtering
	filters := make(map[string]string)

	// Common filter parameters based on the Node.js API
	if name := r.URL.Query().Get("name"); name != "" {
		filters["name"] = name
	}
	if className := r.URL.Query().Get("className"); className != "" {
		filters["className"] = className
	}
	if section := r.URL.Query().Get("section"); section != "" {
		filters["section"] = section
	}
	if roll := r.URL.Query().Get("roll"); roll != "" {
		filters["roll"] = roll
	}

	// Fetch students from the service
	students, err := h.pdfService.GetAllStudents(filters)
	if err != nil {
		statusCode := http.StatusInternalServerError

		// Check if it's a client error (no students found, etc.)
		if isClientError(err) {
			statusCode = http.StatusNotFound
		}

		h.writeErrorResponse(w, statusCode, "Failed to fetch students", err)
		return
	}

	// Return success response
	h.writeSuccessResponse(w, http.StatusOK, "Students retrieved successfully", students)
}

// Helper methods for consistent response formatting

func (h *StudentPDFHandler) writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
	h.writeResponse(w, statusCode, response)
}

func (h *StudentPDFHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	response := ErrorResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err != nil {
		response.Error = err.Error()
	}

	h.writeResponse(w, statusCode, response)
}

func (h *StudentPDFHandler) writeResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If we can't encode the response, write a basic error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Helper function to determine if error is a client error
func isClientError(err error) bool {
	errorStr := err.Error()
	return strings.Contains(errorStr, "not found") ||
		strings.Contains(errorStr, "invalid") ||
		strings.Contains(errorStr, "API Error 4")
}

// Response structures

// APIResponse represents a successful API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
