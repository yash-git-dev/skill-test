package service

import (
	"errors"
	"testing"

	"student-report-service/internal/config"
	"student-report-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNodeJSClient implements NodeJSClientInterface for testing
type MockNodeJSClient struct {
	mock.Mock
}

func (m *MockNodeJSClient) GetStudentByID(studentID int) (*models.Student, error) {
	args := m.Called(studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Student), args.Error(1)
}

func (m *MockNodeJSClient) GetAllStudents(filters map[string]string) ([]models.StudentListItem, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.StudentListItem), args.Error(1)
}

func (m *MockNodeJSClient) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNodeJSClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockPDFGenerator implements PDFGeneratorInterface for testing
type MockPDFGenerator struct {
	mock.Mock
}

func (m *MockPDFGenerator) GenerateStudentReport(student *models.Student, metadata *models.ReportMetadata) (string, error) {
	args := m.Called(student, metadata)
	return args.String(0), args.Error(1)
}

func (m *MockPDFGenerator) CleanupOldReports() error {
	args := m.Called()
	return args.Error(0)
}

func TestPDFReportService_CreateStudentPDF(t *testing.T) {
	mockStudent := &models.Student{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	tests := []struct {
		name          string
		studentID     int
		generatedBy   string
		setupMocks    func(*MockNodeJSClient, *MockPDFGenerator)
		expectedError bool
		errorContains string
	}{
		{
			name:        "Successful report generation",
			studentID:   1,
			generatedBy: "Test User",
			setupMocks: func(nodeClient *MockNodeJSClient, pdfGen *MockPDFGenerator) {
				nodeClient.On("GetStudentByID", 1).Return(mockStudent, nil)
				pdfGen.On("GenerateStudentReport", mock.AnythingOfType("*models.Student"), mock.AnythingOfType("*models.ReportMetadata")).Return("/path/to/report.pdf", nil)
			},
			expectedError: false,
		},
		{
			name:        "Invalid student ID",
			studentID:   0,
			generatedBy: "Test User",
			setupMocks: func(nodeClient *MockNodeJSClient, pdfGen *MockPDFGenerator) {
				// No setup needed - validation happens before API call
			},
			expectedError: true,
			errorContains: "invalid student ID",
		},
		{
			name:        "Student not found",
			studentID:   999,
			generatedBy: "Test User",
			setupMocks: func(nodeClient *MockNodeJSClient, pdfGen *MockPDFGenerator) {
				nodeClient.On("GetStudentByID", 999).Return(nil, errors.New("student not found"))
			},
			expectedError: true,
			errorContains: "failed to fetch student data",
		},
		{
			name:        "PDF generation fails",
			studentID:   1,
			generatedBy: "Test User",
			setupMocks: func(nodeClient *MockNodeJSClient, pdfGen *MockPDFGenerator) {
				nodeClient.On("GetStudentByID", 1).Return(mockStudent, nil)
				pdfGen.On("GenerateStudentReport", mock.AnythingOfType("*models.Student"), mock.AnythingOfType("*models.ReportMetadata")).Return("", errors.New("PDF generation failed"))
			},
			expectedError: true,
			errorContains: "failed to generate PDF report",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockNodeClient := new(MockNodeJSClient)
			mockPDFGen := new(MockPDFGenerator)

			// Setup mocks
			tt.setupMocks(mockNodeClient, mockPDFGen)

			// Create service
			cfg := &config.Config{}
			service := NewPDFReportService(mockNodeClient, mockPDFGen, cfg)

			// Execute
			result, err := service.CreateStudentPDF(tt.studentID, tt.generatedBy)

			// Verify
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify result fields
				assert.Equal(t, tt.studentID, result.StudentID)
				assert.Equal(t, tt.generatedBy, result.GeneratedBy)
				assert.Equal(t, mockStudent.Name, result.StudentName)
			}

			// Assert that all expectations were met
			mockNodeClient.AssertExpectations(t)
			mockPDFGen.AssertExpectations(t)
		})
	}
}

func TestPDFReportService_HealthCheck(t *testing.T) {
	tests := []struct {
		name            string
		setupMocks      func(*MockNodeJSClient, *MockPDFGenerator)
		expectedHealthy bool
	}{
		{
			name: "All components healthy",
			setupMocks: func(nodeClient *MockNodeJSClient, pdfGen *MockPDFGenerator) {
				nodeClient.On("HealthCheck").Return(nil)
			},
			expectedHealthy: true,
		},
		{
			name: "Node.js API unhealthy",
			setupMocks: func(nodeClient *MockNodeJSClient, pdfGen *MockPDFGenerator) {
				nodeClient.On("HealthCheck").Return(errors.New("API unavailable"))
			},
			expectedHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockNodeClient := new(MockNodeJSClient)
			mockPDFGen := new(MockPDFGenerator)

			// Setup mocks
			tt.setupMocks(mockNodeClient, mockPDFGen)

			// Create service
			cfg := &config.Config{}
			service := NewPDFReportService(mockNodeClient, mockPDFGen, cfg)

			// Execute
			status := service.HealthCheck()

			// Verify
			assert.Equal(t, tt.expectedHealthy, status.Healthy)
			assert.Equal(t, "Report Service", status.Service)
			assert.NotEmpty(t, status.Components)

			// Assert that all expectations were met
			mockNodeClient.AssertExpectations(t)
			mockPDFGen.AssertExpectations(t)
		})
	}
}

func TestPDFReportService_CleanupOldReports(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockPDFGenerator)
		expectedError bool
	}{
		{
			name: "Successful cleanup",
			setupMocks: func(pdfGen *MockPDFGenerator) {
				pdfGen.On("CleanupOldReports").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Cleanup fails",
			setupMocks: func(pdfGen *MockPDFGenerator) {
				pdfGen.On("CleanupOldReports").Return(errors.New("cleanup failed"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockNodeClient := new(MockNodeJSClient)
			mockPDFGen := new(MockPDFGenerator)

			// Setup mocks
			tt.setupMocks(mockPDFGen)

			// Create service
			cfg := &config.Config{}
			service := NewPDFReportService(mockNodeClient, mockPDFGen, cfg)

			// Execute
			err := service.CleanupOldReports()

			// Verify
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Assert that all expectations were met
			mockNodeClient.AssertExpectations(t)
			mockPDFGen.AssertExpectations(t)
		})
	}
}

func TestPDFReportService_GetAllStudents(t *testing.T) {
	mockStudents := []models.StudentListItem{
		{
			ID:      1,
			Name:    "John Doe",
			Email:   "john@example.com",
			Class:   stringPtr("Grade 10"),
			Section: stringPtr("A"),
			Roll:    intPtr(101),
		},
		{
			ID:      2,
			Name:    "Jane Smith",
			Email:   "jane@example.com",
			Class:   stringPtr("Grade 10"),
			Section: stringPtr("B"),
			Roll:    intPtr(102),
		},
	}

	tests := []struct {
		name          string
		filters       map[string]string
		setupMocks    func(*MockNodeJSClient)
		expectedError bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "Successful retrieval with no filters",
			filters: map[string]string{},
			setupMocks: func(nodeClient *MockNodeJSClient) {
				nodeClient.On("GetAllStudents", map[string]string{}).Return(mockStudents, nil)
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:    "Successful retrieval with filters",
			filters: map[string]string{"className": "Grade 10", "section": "A"},
			setupMocks: func(nodeClient *MockNodeJSClient) {
				filters := map[string]string{"className": "Grade 10", "section": "A"}
				filteredStudents := []models.StudentListItem{mockStudents[0]}
				nodeClient.On("GetAllStudents", filters).Return(filteredStudents, nil)
			},
			expectedError: false,
			expectedCount: 1,
		},
		{
			name:    "API error",
			filters: map[string]string{},
			setupMocks: func(nodeClient *MockNodeJSClient) {
				nodeClient.On("GetAllStudents", map[string]string{}).Return(nil, errors.New("API unavailable"))
			},
			expectedError: true,
			errorContains: "failed to fetch students list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockNodeClient := new(MockNodeJSClient)
			mockPDFGen := new(MockPDFGenerator)

			// Setup mocks
			tt.setupMocks(mockNodeClient)

			// Create service
			cfg := &config.Config{}
			service := NewPDFReportService(mockNodeClient, mockPDFGen, cfg)

			// Execute
			students, err := service.GetAllStudents(tt.filters)

			// Verify
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, students)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, students)
				assert.Len(t, students, tt.expectedCount)
			}

			// Assert that all expectations were met
			mockNodeClient.AssertExpectations(t)
			mockPDFGen.AssertExpectations(t)
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
