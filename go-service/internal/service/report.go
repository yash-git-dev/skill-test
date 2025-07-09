package service

import (
	"fmt"
	"os"
	"time"

	"student-report-service/internal/client"
	"student-report-service/internal/config"
	"student-report-service/internal/models"
	"student-report-service/internal/pdf"
)

// PDFReportService orchestrates the student report generation process
type PDFReportService struct {
	nodeClient   NodeJSClientInterface
	pdfGenerator PDFGeneratorInterface
	config       *config.Config
}

// NewPDFReportService creates a new report service
func NewPDFReportService(nodeClient NodeJSClientInterface, pdfGenerator PDFGeneratorInterface, cfg *config.Config) *PDFReportService {
	return &PDFReportService{
		nodeClient:   nodeClient,
		pdfGenerator: pdfGenerator,
		config:       cfg,
	}
}

// NewPDFReportServiceWithConcreteTypes creates a new report service with concrete types (for production use)
func NewPDFReportServiceWithConcreteTypes(nodeClient *client.NodeJSClient, pdfGenerator *pdf.Generator, cfg *config.Config) *PDFReportService {
	return &PDFReportService{
		nodeClient:   nodeClient,
		pdfGenerator: pdfGenerator,
		config:       cfg,
	}
}

// GetAllStudents retrieves a list of all students with optional filtering
func (ps *PDFReportService) GetAllStudents(filters map[string]string) ([]models.StudentListItem, error) {
	students, err := ps.nodeClient.GetAllStudents(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch students list: %w", err)
	}

	return students, nil
}

// CreateStudentPDF generates a complete student report
func (ps *PDFReportService) CreateStudentPDF(studentID int, generatedBy string) (*PDFReportResult, error) {
	if studentID <= 0 {
		return nil, fmt.Errorf("invalid student ID: %d", studentID)
	}

	// Step 1: Fetch student data from Node.js API
	student, err := ps.nodeClient.GetStudentByID(studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch student data: %w", err)
	}

	if student == nil {
		return nil, fmt.Errorf("student with ID %d not found", studentID)
	}

	// Step 2: Create report metadata
	metadata := &models.ReportMetadata{
		GeneratedAt: time.Now(),
		GeneratedBy: generatedBy,
		ReportID:    fmt.Sprintf("RPT-%d-%d", studentID, time.Now().Unix()),
	}

	// Step 3: Generate PDF report
	filePath, err := ps.pdfGenerator.GenerateStudentReport(student, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF report: %w", err)
	}

	// Step 4: Get actual file size
	fileSize := ps.getActualFileSize(filePath)

	// Step 5: Create result
	result := &PDFReportResult{
		ReportID:    metadata.ReportID,
		StudentID:   studentID,
		StudentName: student.FormatName(),
		FilePath:    filePath,
		GeneratedAt: metadata.GeneratedAt,
		GeneratedBy: generatedBy,
		FileSize:    fileSize,
	}

	return result, nil
}

// HealthCheck performs a comprehensive health check
func (ps *PDFReportService) HealthCheck() *ServiceHealthStatus {
	status := &ServiceHealthStatus{
		Service:    "Report Service",
		Timestamp:  time.Now(),
		Healthy:    true,
		Components: make(map[string]ComponentStatus),
	}

	// Check Node.js API connectivity
	if err := ps.nodeClient.HealthCheck(); err != nil {
		status.Healthy = false
		status.Components["nodejs_api"] = ComponentStatus{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	} else {
		status.Components["nodejs_api"] = ComponentStatus{
			Status:  "healthy",
			Message: "API is responsive",
		}
	}

	// Check PDF generator (output directory)
	if generator := ps.pdfGenerator; generator != nil {
		status.Components["pdf_generator"] = ComponentStatus{
			Status:  "healthy",
			Message: "Generator is ready",
		}
	} else {
		status.Healthy = false
		status.Components["pdf_generator"] = ComponentStatus{
			Status:  "unhealthy",
			Message: "Generator not initialized",
		}
	}

	// Set overall status message
	if status.Healthy {
		status.Message = "All systems operational"
	} else {
		status.Message = "Some components are unhealthy"
	}

	return status
}

// CleanupOldReports cleans up old report files
func (ps *PDFReportService) CleanupOldReports() error {
	return ps.pdfGenerator.CleanupOldReports()
}

// getActualFileSize gets the actual file size for the generated report
func (ps *PDFReportService) getActualFileSize(filePath string) int64 {
	if fileInfo, err := os.Stat(filePath); err == nil {
		return fileInfo.Size()
	}
	return 0
}

// PDFReportResult represents the result of a report generation
type PDFReportResult struct {
	ReportID    string    `json:"report_id"`
	StudentID   int       `json:"student_id"`
	StudentName string    `json:"student_name"`
	FilePath    string    `json:"file_path"`
	GeneratedAt time.Time `json:"generated_at"`
	GeneratedBy string    `json:"generated_by"`
	FileSize    int64     `json:"file_size"`
}

// ServiceHealthStatus represents the health status of the service
type ServiceHealthStatus struct {
	Service    string                     `json:"service"`
	Healthy    bool                       `json:"healthy"`
	Message    string                     `json:"message"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentStatus `json:"components"`
}

// ComponentStatus represents the status of an individual component
type ComponentStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
