package jobs

import (
	"bytes"
	"fmt"
	"html/template"
	"net/mail"
	"os"
	"path/filepath"
	"time"

	"ims-pocketbase-baas-starter/pkg/cronutils"
	"ims-pocketbase-baas-starter/pkg/jobutils"
	"ims-pocketbase-baas-starter/pkg/logger"
	"ims-pocketbase-baas-starter/pkg/metrics"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

// EmailJobHandler handles email job processing
type EmailJobHandler struct {
	app *pocketbase.PocketBase
}

// NewEmailJobHandler creates a new email job handler
func NewEmailJobHandler(app *pocketbase.PocketBase) *EmailJobHandler {
	return &EmailJobHandler{
		app: app,
	}
}

// Handle processes an email job using typed payload structures (with metrics instrumentation)
func (h *EmailJobHandler) Handle(ctx *cronutils.CronExecutionContext, job *jobutils.JobData) error {
	ctx.LogStart(fmt.Sprintf("Processing email job: %s", job.ID))

	// Get the metrics provider instance
	metricsProvider := metrics.GetInstance()

	// Instrument the job handler execution with metrics collection
	return metrics.InstrumentJobHandler(metricsProvider, "email_job", func() error {
		// Parse payload using centralized utility function
		emailPayload, err := jobutils.ParseEmailJobPayload(job)
		if err != nil {
			return fmt.Errorf("failed to parse email job payload: %w", err)
		}

		// Additional validation specific to this handler
		if err := h.validateEmailPayload(emailPayload); err != nil {
			return fmt.Errorf("invalid email job payload: %w", err)
		}

		// Log email processing details
		logger := logger.GetLogger(h.app)
		logger.Info("Processing email job",
			"job_id", job.ID,
			"to", emailPayload.Data.To,
			"subject", emailPayload.Data.Subject,
			"template", emailPayload.Data.Template,
			"attempts", job.Attempts,
			"retry_count", emailPayload.Options.RetryCount)

		// Process email templates
		htmlContent, textContent, err := h.processEmailTemplates(emailPayload)
		if err != nil {
			return fmt.Errorf("failed to process email templates: %w", err)
		}

		// Send email using PocketBase mailer
		if err := h.sendEmail(emailPayload, htmlContent, textContent); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		// Create result data
		result := &jobutils.EmailResult{
			BaseJobResultData: jobutils.BaseJobResultData{
				Message:   "Email sent successfully",
				Timestamp: time.Now(),
			},
			MessageId:   fmt.Sprintf("msg_%s_%d", job.ID, time.Now().Unix()),
			DeliveredAt: func() *time.Time { t := time.Now(); return &t }(),
			Recipients:  []string{emailPayload.Data.To},
		}

		// Log result for debugging
		ctx.LogDebug(result, "Email job result")

		ctx.LogEnd("Email job processed successfully")
		return nil
	})
}

// GetJobType returns the job type this handler processes
func (h *EmailJobHandler) GetJobType() string {
	return jobutils.JobTypeEmail
}

// validateEmailPayload validates the typed email job payload (additional handler-specific validation)
func (h *EmailJobHandler) validateEmailPayload(payload *jobutils.EmailJobPayload) error {
	// Validate job type matches what this handler expects
	if payload.Type != jobutils.JobTypeEmail {
		return fmt.Errorf("invalid job type: expected %s, got %s", jobutils.JobTypeEmail, payload.Type)
	}

	// Additional validation can be added here for handler-specific requirements
	// Basic validation is already done in the ParseEmailJobPayload function

	return nil
}

// processEmailTemplates processes both HTML and text email templates with variables
func (h *EmailJobHandler) processEmailTemplates(payload *jobutils.EmailJobPayload) (string, string, error) {
	logger := logger.GetLogger(h.app)

	// If no template is specified, return empty content
	if payload.Data.Template == "" {
		logger.Debug("No template specified, using empty content")
		return "", "", nil
	}

	// Try to process HTML template
	htmlContent, err := h.processSingleTemplate(payload, ".html")
	if err != nil {
		logger.Warn("Failed to process HTML template", "error", err)
	}

	// Try to process text template
	textContent, err := h.processSingleTemplate(payload, ".txt")
	if err != nil {
		logger.Warn("Failed to process text template", "error", err)
	}

	logger.Debug("Email templates processed successfully", "template", payload.Data.Template)
	return htmlContent, textContent, nil
}

// processSingleTemplate processes a single email template with variables
func (h *EmailJobHandler) processSingleTemplate(payload *jobutils.EmailJobPayload, extension string) (string, error) {
	// Construct template path
	templatePath := filepath.Join("templates", "emails", payload.Data.Template+extension)

	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}

	// Parse template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template with variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, payload.Data.Variables); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// sendEmail sends the email using PocketBase mailer
func (h *EmailJobHandler) sendEmail(payload *jobutils.EmailJobPayload, htmlContent, textContent string) error {
	logger := logger.GetLogger(h.app)

	// Get sender settings from PocketBase admin UI configuration
	settings := h.app.Settings()

	// Use the configured sender name and address from admin UI
	fromEmail := settings.Meta.SenderAddress
	fromName := settings.Meta.SenderName

	// Fallback to environment variables if admin UI settings are empty
	if fromEmail == "" {
		fromEmail = os.Getenv("SMTP_FROM_EMAIL")
		if fromEmail == "" {
			fromEmail = "noreply@ims-app.local"
		}
	}

	if fromName == "" {
		fromName = os.Getenv("SMTP_FROM_NAME")
		if fromName == "" {
			fromName = "IMS PocketBase App"
		}
	}

	// Create new mailer message
	message := &mailer.Message{
		From:    mail.Address{Name: fromName, Address: fromEmail},
		To:      []mail.Address{{Address: payload.Data.To}},
		Subject: payload.Data.Subject,
	}

	// Set HTML content if available
	if htmlContent != "" {
		message.HTML = htmlContent
	}

	// Set text content if available
	if textContent != "" {
		message.Text = textContent
	} else if htmlContent != "" {
		// If only HTML content is available, use it as text as fallback
		message.Text = htmlContent
	}

	// Send email
	if err := h.app.NewMailClient().Send(message); err != nil {
		logger.Error("Failed to send email",
			"to", payload.Data.To,
			"subject", payload.Data.Subject,
			"error", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Info("Email sent successfully",
		"to", payload.Data.To,
		"subject", payload.Data.Subject)

	return nil
}
