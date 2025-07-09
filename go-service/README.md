# Student Report Generation Microservice

A high-performance Go microservice for generating PDF reports of student information by consuming the Node.js backend API.

## 🚀 Features

- **Clean Architecture**: Follows Domain-Driven Design principles with clear separation of concerns
- **PDF Generation**: Creates professional, formatted PDF reports with student information
- **API Integration**: Consumes Node.js backend API with resty HTTP client, retry logic and error handling
- **Health Monitoring**: Built-in health checks for all components
- **Comprehensive Logging**: Structured logging with configurable levels
- **Graceful Shutdown**: Proper resource cleanup and shutdown handling
- **Configuration Management**: Environment-based configuration with sensible defaults
- **Error Handling**: Robust error handling with appropriate HTTP status codes
- **Testing**: Comprehensive unit tests with mocks and interfaces

## 🏗️ Architecture

```text
go-service/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── client/
│   │   └── nodejs.go          # Node.js API client
│   ├── config/
│   │   └── config.go          # Configuration management
│   ├── handlers/
│   │   └── handlers.go        # HTTP request handlers
│   ├── models/
│   │   ├── student.go         # Data models
│   │   └── student_test.go    # Model tests
│   ├── pdf/
│   │   └── generator.go       # PDF generation logic
│   ├── service/
│   │   ├── report.go          # Business logic layer
│   │   └── report_test.go     # Service tests
├── reports/                   # Generated PDF output directory
├── go.mod                     # Go module definition
└── README.md                  # This file
```

## 🔧 Configuration

The service uses environment variables for configuration with sensible defaults:

### Server Configuration

- `GO_SERVICE_PORT`: Server port (default: 8080)
- `READ_TIMEOUT`: HTTP read timeout (default: 10s)
- `WRITE_TIMEOUT`: HTTP write timeout (default: 10s)
- `IDLE_TIMEOUT`: HTTP idle timeout (default: 60s)

### Node.js API Configuration

- `NODEJS_API_URL`: Base URL for Node.js API (default: <http://localhost:5007/api/v1>)
- `NODEJS_TIMEOUT`: Request timeout (default: 30s)
- `NODEJS_RETRY_ATTEMPTS`: Number of retry attempts (default: 3)
- `NODEJS_RETRY_DELAY`: Delay between retries (default: 1s)

### Report Configuration

- `REPORT_OUTPUT_DIR`: Output directory for PDF files (default: ./reports)
- `REPORT_MAX_FILE_SIZE`: Maximum PDF file size in bytes (default: 10MB)
- `REPORT_CLEANUP`: Enable automatic cleanup (default: true)
- `REPORT_CLEANUP_AFTER`: Cleanup files older than (default: 24h)
- `REPORT_WATERMARK`: Watermark text for PDFs (default: "Student Management System - Confidential")

### Logging Configuration

- `LOG_LEVEL`: Log level (default: info)
- `LOG_FORMAT`: Log format - json or text (default: json)

## 📦 Installation & Setup

### Prerequisites

- Go 1.21 or higher
- Node.js backend service running on port 5007

### Dependencies

Key libraries used in this project:

- **github.com/go-resty/resty/v2**: Modern HTTP client with retry logic and easy JSON handling
- **github.com/gorilla/mux**: HTTP router and URL matcher
- **github.com/jung-kurt/gofpdf**: PDF generation library
- **github.com/rs/cors**: CORS middleware for HTTP handlers
- **github.com/sirupsen/logrus**: Structured logger
- **github.com/stretchr/testify**: Testing toolkit with mocks and assertions

### 1. Install Dependencies

```bash
cd go-service
go mod tidy
```

### 2. Set Environment Variables (Optional)

```bash
export GO_SERVICE_PORT=8080
export NODEJS_API_URL=http://localhost:5007/api/v1
export LOG_LEVEL=debug
```

### 3. Run the Service

```bash
# Development mode
go run cmd/main.go

# Build and run
go build -o student-report-service cmd/main.go
./student-report-service
```

The service will start on <http://localhost:8080>

## 📚 API Documentation

### Health Check

**GET** `/health`

Returns the health status of the service and its dependencies.

**Response:**

```json
{
  "service": "Report Service",
  "healthy": true,
  "message": "All systems operational",
  "timestamp": "2024-01-15T10:30:00Z",
  "components": {
    "nodejs_api": {
      "status": "healthy",
      "message": "API is responsive"
    },
    "pdf_generator": {
      "status": "healthy", 
      "message": "Generator is ready"
    }
  }
}
```

### List Students

**GET** `/api/v1/students`

Retrieves a list of all students with optional filtering.

**Query Parameters:**

- `name` (optional): Filter by student name
- `className` (optional): Filter by class name
- `section` (optional): Filter by section
- `roll` (optional): Filter by roll number

**Example Request:**

```bash
# Get all students
curl "http://localhost:8080/api/v1/students"

# Get students in a specific class
curl "http://localhost:8080/api/v1/students?className=Grade%2010"

# Get students in a specific section
curl "http://localhost:8080/api/v1/students?section=A"
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Students retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john.doe@student.school.com",
      "systemAccess": true,
      "class": "Grade 10",
      "section": "A",
      "roll": 101
    },
    {
      "id": 2,
      "name": "Jane Smith",
      "email": "jane.smith@student.school.com",
      "systemAccess": true,
      "class": "Grade 10",
      "section": "B",
      "roll": 102
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error Response (404 - No Students Found):**

```json
{
  "success": false,
  "message": "Failed to fetch students",
  "error": "API Error 404: Students not found",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Generate Student Report

**POST** `/api/v1/reports/student/{id}`

Generates a PDF report for the specified student ID.

**Parameters:**

- `id` (path): Student ID (integer, required)
- `generated_by` (query): Name of the user generating the report (optional, defaults to "API")

**Example Request:**

```bash
curl -X POST "http://localhost:8080/api/v1/reports/student/123?generated_by=Admin User"
```

**Success Response (201):**

```json
{
  "success": true,
  "message": "Report generated successfully",
  "data": {
    "report_id": "RPT-123-1705312200",
    "student_id": 123,
    "student_name": "John Doe",
    "file_path": "/path/to/student_report_123_John_Doe_20240115_103000.pdf",
    "generated_at": "2024-01-15T10:30:00Z",
    "generated_by": "Admin User",
    "file_size": 245760
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error Response (404 - Student Not Found):**

```json
{
  "success": false,
  "message": "Failed to generate report",
  "error": "API Error 404: Student not found",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Error Response (400 - Invalid ID):**

```json
{
  "success": false,
  "message": "Invalid student ID format",
  "error": "strconv.Atoi: parsing \"abc\": invalid syntax",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Cleanup Old Reports

**POST** `/api/v1/reports/cleanup`

Removes old PDF report files based on the configured cleanup policy.

**Example Request:**

```bash
curl -X POST "http://localhost:8080/api/v1/reports/cleanup"
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Old reports cleaned up successfully",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 🧪 Testing

### Run Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test package
go test ./internal/models
go test ./internal/service
```

### Test Coverage

The service includes comprehensive unit tests for:

- Models and utility functions
- Service layer business logic
- Error handling scenarios
- Mock implementations for external dependencies

## 🚦 Usage Examples

### Basic Usage

1. **Start the Node.js backend** (ensure it's running on port 5007)
2. **Start the Go service**:

   ```bash
   go run cmd/main.go
   ```

3. **Generate a report**:

   ```bash
   curl -X POST "http://localhost:8080/api/v1/reports/student/1"
   ```

### Integration with Frontend

```javascript
// Generate report for student ID 123
const generateReport = async (studentId, generatedBy = 'Frontend User') => {
  try {
    const response = await fetch(
      `http://localhost:8080/api/v1/reports/student/${studentId}?generated_by=${encodeURIComponent(generatedBy)}`,
      { method: 'POST' }
    );
    
    const result = await response.json();
    
    if (result.success) {
      console.log('Report generated:', result.data.file_path);
      return result.data;
    } else {
      throw new Error(result.message);
    }
  } catch (error) {
    console.error('Report generation failed:', error);
    throw error;
  }
};

// Usage
generateReport(123, 'Admin User')
  .then(report => console.log('Report ID:', report.report_id))
  .catch(error => console.error('Error:', error));
```

## 📊 Generated PDF Features

The generated PDF reports include:

- **Header Section**: Report title, metadata, and watermark
- **Basic Information**: Student ID, name, email, system access status
- **Contact Information**: Phone number and email details
- **Family Information**: Father, mother, and guardian details with contact information
- **Address Information**: Current and permanent addresses
- **Academic Information**: Class, section, roll number, admission date
- **Footer**: Confidentiality notice and generation timestamp

## 🔒 Security Considerations

- **Input Validation**: All inputs are validated before processing
- **Error Handling**: Sensitive information is not exposed in error messages
- **File Security**: Generated files are stored in a controlled directory
- **CORS Configuration**: Properly configured for production use
- **Watermarking**: All PDFs include confidentiality watermarks

## 🚀 Production Deployment

### Docker Deployment (Recommended)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o student-report-service cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/student-report-service .
EXPOSE 8080
CMD ["./student-report-service"]
```

### Environment Configuration for Production

```bash
export GO_SERVICE_PORT=8080
export NODEJS_API_URL=http://nodejs-api:5007/api/v1
export LOG_LEVEL=info
export LOG_FORMAT=json
export REPORT_OUTPUT_DIR=/app/reports
export REPORT_CLEANUP=true
export REPORT_CLEANUP_AFTER=24h
```

## 🔧 Monitoring & Observability

### Health Checks

- Service health endpoint: `GET /health`
- Monitors Node.js API connectivity
- Checks PDF generator availability
- Returns detailed component status

### Logging

- Structured JSON logging in production
- Request/response logging with timing
- Error logging with stack traces
- Configurable log levels

### Metrics (Future Enhancement)

- Request duration histograms
- Success/error rate counters
- PDF generation performance metrics
- Node.js API response time tracking

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📝 License

This project is part of the Student Management System and follows the same licensing terms.

## 🆘 Troubleshooting

### Common Issues

1. **"Failed to fetch student data"**
   - Ensure Node.js backend is running on the correct port
   - Check NODEJS_API_URL configuration
   - Verify student ID exists in the database

2. **"PDF generation failed"**
   - Check write permissions for REPORT_OUTPUT_DIR
   - Ensure sufficient disk space
   - Verify REPORT_MAX_FILE_SIZE settings

3. **"Connection refused"**
   - Verify Node.js backend is accessible
   - Check network connectivity
   - Review firewall settings

### Debug Mode

```bash
export LOG_LEVEL=debug
go run cmd/main.go
```

This will provide detailed logging for troubleshooting issues.
