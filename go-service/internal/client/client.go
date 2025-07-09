package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"student-report-service/internal/config"
	"student-report-service/internal/models"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// NodeJSClient handles communication with the Node.js backend API
type NodeJSClient struct {
	client  *resty.Client
	config  *config.NodeJSConfig
	logger  *logrus.Logger
	baseURL string

	// Authentication state - manual token management
	accessToken  string
	refreshToken string
	csrfToken    string
	authMutex    sync.RWMutex
}

// ClientError represents errors from the Node.js API
type ClientError struct {
	StatusCode int
	Message    string
	Details    string
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("API Error %d: %s", e.StatusCode, e.Message)
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	// ... other fields are not needed for our use case
}

// NewNodeJSClient creates a new NodeJS API client with authentication support
func NewNodeJSClient(cfg *config.NodeJSConfig, logger *logrus.Logger) (*NodeJSClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	if cfg.ServiceUsername == "" || cfg.ServicePassword == "" {
		logger.Warn("Service credentials not configured - authentication may fail")
	}

	// Create resty client with retry configuration
	client := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.RetryAttempts).
		SetRetryWaitTime(cfg.RetryDelay).
		SetRetryMaxWaitTime(cfg.RetryDelay*5).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	// Enable debug logging if logger level is debug
	if logger.Level == logrus.DebugLevel {
		client.SetDebug(true)
	}

	return &NodeJSClient{
		client:  client,
		config:  cfg,
		logger:  logger,
		baseURL: cfg.BaseURL,
	}, nil
}

// authenticate performs login and stores authentication tokens
func (c *NodeJSClient) authenticate() error {
	c.logger.WithFields(logrus.Fields{
		"username": c.config.ServiceUsername,
		"base_url": c.baseURL,
	}).Debug("Authenticating with Node.js API")

	loginReq := LoginRequest{
		Username: c.config.ServiceUsername,
		Password: c.config.ServicePassword,
	}

	var loginResp LoginResponse
	var errorResp models.ErrorResponse

	resp, err := c.client.R().
		SetResult(&loginResp).
		SetError(&errorResp).
		SetBody(loginReq).
		Post("/auth/login")

	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	if resp.IsError() {
		if errorResp.Message != "" {
			return &ClientError{
				StatusCode: resp.StatusCode(),
				Message:    errorResp.Message,
				Details:    errorResp.Error,
			}
		}
		return &ClientError{
			StatusCode: resp.StatusCode(),
			Message:    resp.Status(),
			Details:    string(resp.Body()),
		}
	}

	// Extract tokens from Set-Cookie headers in the response
	c.authMutex.Lock()
	defer c.authMutex.Unlock()

	var accessToken, refreshToken, csrfToken string

	// Parse Set-Cookie headers from the response
	setCookieHeaders := resp.Header().Values("Set-Cookie")
	c.logger.WithField("set_cookie_headers", setCookieHeaders).Debug("Received Set-Cookie headers")

	for _, cookieHeader := range setCookieHeaders {
		c.logger.WithField("cookie_header", cookieHeader).Debug("Processing cookie header")

		// Parse each Set-Cookie header
		if strings.HasPrefix(cookieHeader, "accessToken=") {
			// Extract value between accessToken= and the first semicolon
			parts := strings.Split(cookieHeader, ";")
			if len(parts) > 0 {
				tokenPart := strings.TrimPrefix(parts[0], "accessToken=")
				if tokenPart != "" && tokenPart != "accessToken=" {
					accessToken = tokenPart
					c.logger.WithField("access_token_length", len(accessToken)).Debug("Extracted access token")
				}
			}
		} else if strings.HasPrefix(cookieHeader, "refreshToken=") {
			parts := strings.Split(cookieHeader, ";")
			if len(parts) > 0 {
				tokenPart := strings.TrimPrefix(parts[0], "refreshToken=")
				if tokenPart != "" && tokenPart != "refreshToken=" {
					refreshToken = tokenPart
					c.logger.WithField("refresh_token_length", len(refreshToken)).Debug("Extracted refresh token")
				}
			}
		} else if strings.HasPrefix(cookieHeader, "csrfToken=") {
			parts := strings.Split(cookieHeader, ";")
			if len(parts) > 0 {
				tokenPart := strings.TrimPrefix(parts[0], "csrfToken=")
				if tokenPart != "" && tokenPart != "csrfToken=" {
					csrfToken = tokenPart
					c.logger.WithField("csrf_token", csrfToken).Debug("Extracted CSRF token")
				}
			}
		}
	}

	if accessToken == "" || refreshToken == "" || csrfToken == "" {
		return fmt.Errorf("failed to extract authentication tokens: accessToken=%t, refreshToken=%t, csrfToken=%t",
			accessToken != "", refreshToken != "", csrfToken != "")
	}

	// Store tokens for subsequent requests
	c.accessToken = accessToken
	c.refreshToken = refreshToken
	c.csrfToken = csrfToken

	c.logger.WithFields(logrus.Fields{
		"access_token_length":  len(c.accessToken),
		"refresh_token_length": len(c.refreshToken),
		"csrf_token":           c.csrfToken,
	}).Debug("Successfully authenticated with Node.js API")

	return nil
}

// ensureAuthenticated ensures we have valid authentication tokens
func (c *NodeJSClient) ensureAuthenticated() error {
	c.authMutex.RLock()
	hasTokens := c.accessToken != "" && c.refreshToken != "" && c.csrfToken != ""
	c.authMutex.RUnlock()

	if !hasTokens {
		return c.authenticate()
	}
	return nil
}

// makeAuthenticatedRequest makes a request with authentication headers
func (c *NodeJSClient) makeAuthenticatedRequest(method, endpoint string) (*resty.Response, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	c.authMutex.RLock()
	accessToken := c.accessToken
	refreshToken := c.refreshToken
	csrfToken := c.csrfToken
	c.authMutex.RUnlock()

	var errorResp models.ErrorResponse

	// Use manual cookie headers for reliable authentication - don't set result here
	resp, err := c.client.R().
		SetError(&errorResp).
		SetHeader("X-CSRF-TOKEN", csrfToken).
		SetHeader("Cookie", fmt.Sprintf("accessToken=%s; refreshToken=%s", accessToken, refreshToken)).
		Execute(method, endpoint)

	// If we get a 401, try to re-authenticate once
	if err == nil && resp.StatusCode() == 401 {
		c.logger.Debug("Received 401, attempting to re-authenticate")
		if authErr := c.authenticate(); authErr != nil {
			return resp, fmt.Errorf("re-authentication failed: %w", authErr)
		}

		// Retry the request with new tokens
		c.authMutex.RLock()
		newAccessToken := c.accessToken
		newRefreshToken := c.refreshToken
		newCsrfToken := c.csrfToken
		c.authMutex.RUnlock()

		resp, err = c.client.R().
			SetError(&errorResp).
			SetHeader("X-CSRF-TOKEN", newCsrfToken).
			SetHeader("Cookie", fmt.Sprintf("accessToken=%s; refreshToken=%s", newAccessToken, newRefreshToken)).
			Execute(method, endpoint)
	}

	return resp, err
}

// GetStudentByID retrieves a student by ID from the Node.js API with authentication
func (c *NodeJSClient) GetStudentByID(studentID int) (*models.Student, error) {
	if studentID <= 0 {
		return nil, fmt.Errorf("invalid student ID: %d", studentID)
	}

	endpoint := fmt.Sprintf("/students/%d", studentID)

	c.logger.WithFields(logrus.Fields{
		"student_id": studentID,
		"endpoint":   endpoint,
	}).Debug("Making authenticated request to Node.js API")

	resp, err := c.makeAuthenticatedRequest("GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Log the response
	c.logger.WithFields(logrus.Fields{
		"status_code": resp.StatusCode(),
		"body_size":   len(resp.Body()),
	}).Debug("Received response from Node.js API")

	// Check for HTTP errors
	if resp.IsError() {
		var errorResp models.ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil && errorResp.Message != "" {
			return nil, &ClientError{
				StatusCode: resp.StatusCode(),
				Message:    errorResp.Message,
				Details:    errorResp.Error,
			}
		}

		return nil, &ClientError{
			StatusCode: resp.StatusCode(),
			Message:    resp.Status(),
			Details:    string(resp.Body()),
		}
	}

	var apiResp models.APIResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		return nil, &ClientError{
			StatusCode: resp.StatusCode(),
			Message:    apiResp.Message,
			Details:    "API returned success=false",
		}
	}

	return &apiResp.Data, nil
}

// GetAllStudents retrieves all students from the Node.js API with optional filtering
func (c *NodeJSClient) GetAllStudents(filters map[string]string) ([]models.StudentListItem, error) {
	endpoint := "/students"

	// Build query parameters
	if len(filters) > 0 {
		queryParams := make([]string, 0, len(filters))
		for key, value := range filters {
			if value != "" {
				queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
			}
		}
		if len(queryParams) > 0 {
			endpoint += "?" + strings.Join(queryParams, "&")
		}
	}

	c.logger.WithFields(logrus.Fields{
		"endpoint": endpoint,
		"filters":  filters,
	}).Debug("Making authenticated request to fetch all students")

	resp, err := c.makeAuthenticatedRequest("GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Log the response
	c.logger.WithFields(logrus.Fields{
		"status_code": resp.StatusCode(),
		"body_size":   len(resp.Body()),
	}).Debug("Received students list response from Node.js API")

	// Check for HTTP errors
	if resp.IsError() {
		var errorResp models.ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil && errorResp.Message != "" {
			return nil, &ClientError{
				StatusCode: resp.StatusCode(),
				Message:    errorResp.Message,
				Details:    errorResp.Error,
			}
		}

		return nil, &ClientError{
			StatusCode: resp.StatusCode(),
			Message:    resp.Status(),
			Details:    string(resp.Body()),
		}
	}

	var apiResp models.StudentListResponse
	if err := json.Unmarshal(resp.Body(), &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		return nil, &ClientError{
			StatusCode: resp.StatusCode(),
			Message:    apiResp.Message,
			Details:    "API returned success=false",
		}
	}

	return apiResp.Data, nil
}

// HealthCheck performs a health check against the Node.js API
func (c *NodeJSClient) HealthCheck() error {
	// For health check, we'll use a simple request to the base API URL
	// without authentication to avoid circular dependencies
	healthClient := resty.New().
		SetBaseURL(c.baseURL).
		SetTimeout(5 * time.Second)

	resp, err := healthClient.R().Get("/")

	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}

	if resp.StatusCode() >= 500 {
		return fmt.Errorf("Node.js API is experiencing server errors (status: %d)", resp.StatusCode())
	}

	return nil
}

// Close closes the client and cleans up resources
func (c *NodeJSClient) Close() error {
	// Resty client doesn't need explicit closing, but we can implement cleanup logic here
	c.logger.Debug("NodeJS client closed")
	return nil
}
