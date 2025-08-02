package jobutils

import (
	"encoding/json"
	"fmt"
)

// ParseUserExportJobPayload helper function to parse user export job payload
func ParseUserExportJobPayload(job *JobData) (*UserExportJobPayload, error) {
	if job == nil {
		return nil, fmt.Errorf("job data cannot be nil")
	}

	if job.Payload == nil {
		return nil, fmt.Errorf("job payload cannot be nil")
	}

	var payload UserExportJobPayload

	// Convert the generic payload map back to our typed structure
	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user export payload: %w", err)
	}

	// Validate required fields
	if payload.Type == "" {
		return nil, fmt.Errorf("payload type is required")
	}

	if payload.Data.Format == "" {
		return nil, fmt.Errorf("data format is required")
	}

	if payload.Data.UserID == "" {
		return nil, fmt.Errorf("data user_id is required")
	}

	return &payload, nil
}

// ParseEmailJobPayload helper function to parse email job payload
func ParseEmailJobPayload(job *JobData) (*EmailJobPayload, error) {
	if job == nil {
		return nil, fmt.Errorf("job data cannot be nil")
	}

	if job.Payload == nil {
		return nil, fmt.Errorf("job payload cannot be nil")
	}

	var payload EmailJobPayload

	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal email payload: %w", err)
	}

	// Validate required fields
	if payload.Type == "" {
		return nil, fmt.Errorf("payload type is required")
	}

	if payload.Data.To == "" {
		return nil, fmt.Errorf("data to field is required")
	}

	if payload.Data.Subject == "" {
		return nil, fmt.Errorf("data subject is required")
	}

	return &payload, nil
}

// ParseDataProcessingJobPayload helper function to parse data processing job payload
func ParseDataProcessingJobPayload(job *JobData) (*DataProcessingJobPayload, error) {
	if job == nil {
		return nil, fmt.Errorf("job data cannot be nil")
	}

	if job.Payload == nil {
		return nil, fmt.Errorf("job payload cannot be nil")
	}

	var payload DataProcessingJobPayload

	payloadBytes, err := json.Marshal(job.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data processing payload: %w", err)
	}

	// Validate required fields
	if payload.Type == "" {
		return nil, fmt.Errorf("payload type is required")
	}

	if payload.Data.Operation == "" {
		return nil, fmt.Errorf("data operation is required")
	}

	if payload.Data.Source == "" {
		return nil, fmt.Errorf("data source is required")
	}

	return &payload, nil
}
