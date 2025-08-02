package common

import (
	"io"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// PocketBaseResponse represents the standard PocketBase API response structure
type PocketBaseResponse struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

// ResponseHelper provides utility methods for consistent API responses
type ResponseHelper struct{}

// NewResponseHelper creates a new ResponseHelper instance
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// Success sends a successful response with optional data
func (r *ResponseHelper) Success(e *core.RequestEvent, status int, message string, data map[string]any) error {
	if data == nil {
		data = map[string]any{}
	}

	response := PocketBaseResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	return e.JSON(status, response)
}

// Error sends an error response with optional error details
func (r *ResponseHelper) Error(e *core.RequestEvent, status int, message string, errorDetails map[string]any) error {
	if errorDetails == nil {
		errorDetails = map[string]any{}
	}

	response := PocketBaseResponse{
		Status:  status,
		Message: message,
		Data:    errorDetails,
	}

	return e.JSON(status, response)
}

// Common HTTP status responses

// OK sends a 200 OK response
func (r *ResponseHelper) OK(e *core.RequestEvent, message string, data map[string]any) error {
	return r.Success(e, http.StatusOK, message, data)
}

// Created sends a 201 Created response
func (r *ResponseHelper) Created(e *core.RequestEvent, message string, data map[string]any) error {
	return r.Success(e, http.StatusCreated, message, data)
}

// BadRequest sends a 400 Bad Request response
func (r *ResponseHelper) BadRequest(e *core.RequestEvent, message string, errorDetails map[string]any) error {
	return r.Error(e, http.StatusBadRequest, message, errorDetails)
}

// Unauthorized sends a 401 Unauthorized response
func (r *ResponseHelper) Unauthorized(e *core.RequestEvent, message string) error {
	return r.Error(e, http.StatusUnauthorized, message, nil)
}

// Forbidden sends a 403 Forbidden response
func (r *ResponseHelper) Forbidden(e *core.RequestEvent, message string) error {
	return r.Error(e, http.StatusForbidden, message, nil)
}

// NotFound sends a 404 Not Found response
func (r *ResponseHelper) NotFound(e *core.RequestEvent, message string) error {
	return r.Error(e, http.StatusNotFound, message, nil)
}

// InternalServerError sends a 500 Internal Server Error response
func (r *ResponseHelper) InternalServerError(e *core.RequestEvent, message string, errorDetails map[string]any) error {
	return r.Error(e, http.StatusInternalServerError, message, errorDetails)
}

// ValidationError sends a 400 Bad Request response with validation error details
func (r *ResponseHelper) ValidationError(e *core.RequestEvent, message string, validationErrors map[string]any) error {
	errorData := map[string]any{
		"validation": validationErrors,
	}
	return r.BadRequest(e, message, errorData)
}

// File serves a file download from PocketBase filesystem
func (r *ResponseHelper) File(e *core.RequestEvent, fileName, basePath string) error {

	// Create filesystem instance
	filesystem, err := e.App.NewFilesystem()
	if err != nil {
		return r.InternalServerError(e, "Failed to access filesystem", nil)
	}
	defer filesystem.Close()

	// Get file reader from PocketBase filesystem
	fileKey := basePath + "/" + fileName
	fileReader, err := filesystem.GetReader(fileKey)
	if err != nil {
		return r.NotFound(e, "File not accessible")
	}
	defer fileReader.Close()

	// Set appropriate headers for file download
	e.Response.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	e.Response.Header().Set("Content-Type", "application/octet-stream")

	// Copy file content to response
	_, err = io.Copy(e.Response, fileReader)
	if err != nil {
		return r.InternalServerError(e, "Failed to serve file", nil)
	}

	return nil
}

// Global instance for easy access
var Response = NewResponseHelper()
