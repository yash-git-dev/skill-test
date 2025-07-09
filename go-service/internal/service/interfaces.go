package service

import "student-report-service/internal/models"

// NodeJSClientInterface defines the interface for Node.js API client
type NodeJSClientInterface interface {
	GetStudentByID(studentID int) (*models.Student, error)
	GetAllStudents(filters map[string]string) ([]models.StudentListItem, error)
	HealthCheck() error
	Close() error
}

// PDFGeneratorInterface defines the interface for PDF generation
type PDFGeneratorInterface interface {
	GenerateStudentReport(student *models.Student, metadata *models.ReportMetadata) (string, error)
	CleanupOldReports() error
}
