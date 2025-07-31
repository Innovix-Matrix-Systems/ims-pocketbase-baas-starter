package jobutils

import (
	"testing"
)

func TestParseUserExportJobPayload(t *testing.T) {
	tests := []struct {
		name        string
		jobData     *JobData
		expectError bool
		expected    *UserExportJobPayload
	}{
		{
			name: "valid user export payload",
			jobData: &JobData{
				ID:   "job-123",
				Type: "user_export",
				Payload: map[string]interface{}{
					"type": "user_export",
					"data": map[string]interface{}{
						"format":  "csv",
						"fields":  []interface{}{"name", "email", "verified"},
						"user_id": "user-456",
					},
					"options": map[string]interface{}{
						"filename_prefix": "users_export",
						"store_result":    true,
						"result_expiry":   "24h",
					},
				},
			},
			expectError: false,
			expected: &UserExportJobPayload{
				Type: "user_export",
				Data: UserExportJobData{
					Format: "csv",
					Fields: []string{"name", "email", "verified"},
					UserID: "user-456",
				},
				Options: UserExportJobOptions{
					FilenamePrefix: "users_export",
					StoreResult:    true,
					ResultExpiry:   "24h",
				},
			},
		},
		{
			name: "invalid payload structure",
			jobData: &JobData{
				ID:   "job-123",
				Type: "user_export",
				Payload: map[string]interface{}{
					"invalid": "structure",
				},
			},
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseUserExportJobPayload(tt.jobData)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Type != tt.expected.Type {
				t.Errorf("expected type %s, got %s", tt.expected.Type, result.Type)
			}

			if result.Data.Format != tt.expected.Data.Format {
				t.Errorf("expected format %s, got %s", tt.expected.Data.Format, result.Data.Format)
			}

			if result.Data.UserID != tt.expected.Data.UserID {
				t.Errorf("expected user_id %s, got %s", tt.expected.Data.UserID, result.Data.UserID)
			}

			if len(result.Data.Fields) != len(tt.expected.Data.Fields) {
				t.Errorf("expected %d fields, got %d", len(tt.expected.Data.Fields), len(result.Data.Fields))
			}
		})
	}
}

func TestParseEmailJobPayload(t *testing.T) {
	tests := []struct {
		name        string
		jobData     *JobData
		expectError bool
		expected    *EmailJobPayload
	}{
		{
			name: "valid email payload",
			jobData: &JobData{
				ID:   "job-123",
				Type: "email",
				Payload: map[string]interface{}{
					"type": "email",
					"data": map[string]interface{}{
						"to":       "test@example.com",
						"subject":  "Test Email",
						"template": "welcome",
						"variables": map[string]interface{}{
							"name": "John Doe",
							"code": "123456",
						},
					},
					"options": map[string]interface{}{
						"retry_count": float64(3), // JSON numbers are float64
						"timeout":     float64(30),
					},
				},
			},
			expectError: false,
			expected: &EmailJobPayload{
				Type: "email",
				Data: EmailJobData{
					To:       "test@example.com",
					Subject:  "Test Email",
					Template: "welcome",
					Variables: map[string]interface{}{
						"name": "John Doe",
						"code": "123456",
					},
				},
				Options: EmailJobOptions{
					RetryCount: 3,
					Timeout:    30,
				},
			},
		},
		{
			name: "invalid email payload",
			jobData: &JobData{
				ID:   "job-123",
				Type: "email",
				Payload: map[string]interface{}{
					"invalid": "structure",
				},
			},
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEmailJobPayload(tt.jobData)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Type != tt.expected.Type {
				t.Errorf("expected type %s, got %s", tt.expected.Type, result.Type)
			}

			if result.Data.To != tt.expected.Data.To {
				t.Errorf("expected to %s, got %s", tt.expected.Data.To, result.Data.To)
			}

			if result.Data.Subject != tt.expected.Data.Subject {
				t.Errorf("expected subject %s, got %s", tt.expected.Data.Subject, result.Data.Subject)
			}
		})
	}
}

func TestParseDataProcessingJobPayload(t *testing.T) {
	tests := []struct {
		name        string
		jobData     *JobData
		expectError bool
		expected    *DataProcessingJobPayload
	}{
		{
			name: "valid data processing payload",
			jobData: &JobData{
				ID:   "job-123",
				Type: "data_processing",
				Payload: map[string]interface{}{
					"type": "data_processing",
					"data": map[string]interface{}{
						"operation": "transform",
						"source":    "users",
						"target":    "processed_users",
					},
					"options": map[string]interface{}{
						"timeout": float64(300),
					},
				},
			},
			expectError: false,
			expected: &DataProcessingJobPayload{
				Type: "data_processing",
				Data: DataProcessingJobData{
					Operation: "transform",
					Source:    "users",
					Target:    "processed_users",
				},
				Options: DataProcessingJobOptions{
					Timeout: 300,
				},
			},
		},
		{
			name: "invalid data processing payload",
			jobData: &JobData{
				ID:   "job-123",
				Type: "data_processing",
				Payload: map[string]interface{}{
					"invalid": "structure",
				},
			},
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDataProcessingJobPayload(tt.jobData)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Type != tt.expected.Type {
				t.Errorf("expected type %s, got %s", tt.expected.Type, result.Type)
			}

			if result.Data.Operation != tt.expected.Data.Operation {
				t.Errorf("expected operation %s, got %s", tt.expected.Data.Operation, result.Data.Operation)
			}

			if result.Data.Source != tt.expected.Data.Source {
				t.Errorf("expected source %s, got %s", tt.expected.Data.Source, result.Data.Source)
			}
		})
	}
}

func TestParseJobPayloadWithNilJobData(t *testing.T) {
	// Test all parsing functions with nil input
	_, err1 := ParseUserExportJobPayload(nil)
	_, err2 := ParseEmailJobPayload(nil)
	_, err3 := ParseDataProcessingJobPayload(nil)

	if err1 == nil || err2 == nil || err3 == nil {
		t.Error("expected errors when parsing nil job data")
	}
}

func TestParseJobPayloadWithEmptyPayload(t *testing.T) {
	emptyJobData := &JobData{
		ID:      "job-123",
		Type:    "test",
		Payload: map[string]interface{}{},
	}

	_, err1 := ParseUserExportJobPayload(emptyJobData)
	_, err2 := ParseEmailJobPayload(emptyJobData)
	_, err3 := ParseDataProcessingJobPayload(emptyJobData)

	if err1 == nil || err2 == nil || err3 == nil {
		t.Error("expected errors when parsing empty payload")
	}
}
