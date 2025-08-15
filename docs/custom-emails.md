# Custom Email System Guide

Complete guide for sending custom emails using the job queue system with template support.

## Overview

The email system uses a job queue architecture with template processing, allowing you to send emails asynchronously with custom templates and variables.

## Email System Architecture

- **Job Queue**: Emails are processed asynchronously through the job queue system
- **Templates**: HTML and text templates with Go template syntax
- **Variables**: Dynamic content injection using template variables
- **SMTP Configuration**: Configurable SMTP settings via environment variables

## SMTP Configuration

Configure email settings in your `.env` file:

```bash
# SMTP Configuration (for email notifications)
SMTP_ENABLED=true
SMTP_HOST=mailhog                    # Use mailhog for development
SMTP_PORT=1025                       # Port 1025 for mailhog, 587 for production
SMTP_USERNAME=                       # Leave empty for mailhog
SMTP_PASSWORD=                       # Leave empty for mailhog
SMTP_AUTH_METHOD=PLAIN
SMTP_TLS=false                       # false for mailhog, true for production
SMTP_FROM_EMAIL=noreply@ims-app.local
SMTP_FROM_NAME=IMS PocketBase App
```

### Production SMTP Example

```bash
SMTP_ENABLED=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_AUTH_METHOD=PLAIN
SMTP_TLS=true
SMTP_FROM_EMAIL=noreply@yourapp.com
SMTP_FROM_NAME=Your App Name
```

## Creating Email Templates

### 1. Template Structure

Create both HTML and text versions in `templates/emails/`:

```
templates/
└── emails/
    ├── welcome.html          # HTML version
    ├── welcome.txt           # Text version
    ├── password-reset.html   # Custom template
    └── password-reset.txt    # Text version
```

### 2. Template Variables

Templates use Go template syntax with variables from `EmailJobData.Variables`:

**HTML Template Example** (`templates/emails/welcome.html`):
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to {{.AppName}}</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; }
        .container { background: #fff; padding: 30px; border-radius: 8px; }
        .header { text-align: center; border-bottom: 1px solid #eee; }
        .button { background: #3498db; color: white; padding: 12px 24px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to {{.AppName}}!</h1>
        </div>
        <div class="content">
            <p>Hi {{.Name}},</p>
            <p>Welcome to {{.AppName}}! We're excited to have you on board.</p>
            <p>Your account: <strong>{{.Email}}</strong></p>
            {{if .ActivationLink}}
            <p><a href="{{.ActivationLink}}" class="button">Activate Account</a></p>
            {{end}}
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
```

**Text Template Example** (`templates/emails/welcome.txt`):
```text
Welcome to {{.AppName}}!

Hi {{.Name}},

Welcome to {{.AppName}}! We're excited to have you on board.

Your account has been successfully created with the email: {{.Email}}

{{if .ActivationLink}}
Activate your account: {{.ActivationLink}}
{{end}}

Best regards,
The {{.AppName}} Team

© {{.Year}} {{.AppName}}. All rights reserved.
{{.AppURL}}
```

## Sending Emails

### 1. Via API (HTTP Request)

Send emails by creating job queue records:

```bash
curl -X POST http://localhost:8090/api/collections/queues/records \
  -H "Content-Type: application/json" \
  -d '{
    "name": "welcome_email",
    "description": "Send welcome email to new user",
    "payload": {
      "type": "email",
      "data": {
        "to": "user@example.com",
        "subject": "Welcome to Our App!",
        "template": "welcome",
        "variables": {
          "AppName": "My Application",
          "Name": "John Doe",
          "Email": "user@example.com",
          "Year": "2025",
          "AppURL": "https://myapp.com",
          "ActivationLink": "https://myapp.com/activate?token=abc123"
        }
      },
      "options": {
        "retry_count": 3,
        "timeout": 30
      }
    }
  }'
```

### 2. Programmatically in Go

#### Basic Email Function

```go
package main

import (
    "ims-pocketbase-baas-starter/pkg/jobutils"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

func SendCustomEmail(app *pocketbase.PocketBase, to, subject, template string, variables map[string]any) error {
    // Get the queues collection
    collection, err := app.FindCollectionByNameOrId("queues")
    if err != nil {
        return fmt.Errorf("failed to find queues collection: %w", err)
    }
    
    // Create email job payload
    payload := map[string]interface{}{
        "type": jobutils.JobTypeEmail,
        "data": map[string]interface{}{
            "to":        to,
            "subject":   subject,
            "template":  template,
            "variables": variables,
        },
        "options": map[string]interface{}{
            "retry_count": 3,
            "timeout":     30,
        },
    }
    
    // Create queue record
    record := core.NewRecord(collection)
    record.Set("name", fmt.Sprintf("email_%s", template))
    record.Set("description", fmt.Sprintf("Send %s email to %s", template, to))
    record.Set("payload", payload)
    
    // Save the job to queue
    return app.Save(record)
}
```

#### Welcome Email Helper

```go
func SendWelcomeEmail(app *pocketbase.PocketBase, userEmail, userName string) error {
    variables := map[string]any{
        "AppName": os.Getenv("APP_NAME"),
        "Name":    userName,
        "Email":   userEmail,
        "Year":    time.Now().Format("2006"),
        "AppURL":  os.Getenv("APP_URL"),
    }
    
    return SendCustomEmail(app, userEmail, "Welcome to Our App!", "welcome", variables)
}
```

#### Password Reset Email

```go
func SendPasswordResetEmail(app *pocketbase.PocketBase, userEmail, resetToken string) error {
    resetLink := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("APP_URL"), resetToken)
    
    variables := map[string]any{
        "AppName":   os.Getenv("APP_NAME"),
        "Email":     userEmail,
        "ResetLink": resetLink,
        "Year":      time.Now().Format("2006"),
        "AppURL":    os.Getenv("APP_URL"),
    }
    
    return SendCustomEmail(app, userEmail, "Password Reset Request", "password-reset", variables)
}
```

### 3. Using in Event Hooks

#### User Registration Hook

```go
// In internal/handlers/hook/user_hooks.go
func HandleUserWelcomeEmail(e *core.RecordEvent) error {
    // Only process users collection
    if e.Record.Collection().Name != "users" {
        return e.Next()
    }
    
    userEmail := e.Record.GetString("email")
    userName := e.Record.GetString("name")
    
    if userEmail == "" {
        return e.Next() // Skip if no email
    }
    
    // Send welcome email asynchronously
    if err := SendWelcomeEmail(e.App, userEmail, userName); err != nil {
        // Log error but don't fail the user creation
        if log := logger.FromApp(e.App); log != nil {
            log.Error("Failed to queue welcome email", "error", err, "email", userEmail)
        }
    }
    
    return e.Next()
}
```

## Email Job Processing

### Job Handler Details

The `EmailJobHandler` in `internal/handlers/jobs/email_job_handler.go` processes email jobs:

- **Template Processing**: Loads and processes both HTML and text templates
- **Variable Substitution**: Injects variables into templates using Go template engine
- **SMTP Integration**: Uses PocketBase's mailer with configured SMTP settings
- **Error Handling**: Comprehensive logging and error reporting
- **Retry Logic**: Automatic retry on failure based on job options

### Job Payload Structure

```go
type EmailJobPayload struct {
    Type    string          `json:"type"`     // Must be "email"
    Data    EmailJobData    `json:"data"`     // Email details
    Options EmailJobOptions `json:"options"`  // Processing options
}

type EmailJobData struct {
    To        string         `json:"to"`        // Recipient email
    Subject   string         `json:"subject"`   // Email subject
    Template  string         `json:"template"`  // Template name (without extension)
    Variables map[string]any `json:"variables"` // Template variables
}

type EmailJobOptions struct {
    RetryCount int `json:"retry_count"` // Number of retries on failure
    Timeout    int `json:"timeout"`     // Timeout in seconds
}
```

## Common Email Templates

### 1. Account Activation

**Template**: `account-activation.html`
**Variables**: `Name`, `Email`, `ActivationLink`, `AppName`, `Year`, `AppURL`

### 2. Password Reset

**Template**: `password-reset.html`
**Variables**: `Email`, `ResetLink`, `AppName`, `Year`, `AppURL`

### 3. Email Verification

**Template**: `email-verification.html`
**Variables**: `Name`, `Email`, `VerificationLink`, `AppName`, `Year`, `AppURL`

### 4. Order Confirmation

**Template**: `order-confirmation.html`
**Variables**: `Name`, `OrderId`, `OrderTotal`, `OrderItems`, `AppName`, `Year`, `AppURL`

## Testing Emails

### Development with MailHog

MailHog captures emails in development:

1. **Start MailHog**: Included in `docker-compose.yml`
2. **View Emails**: Visit `http://localhost:8025`
3. **Configuration**: Use the development SMTP settings shown above

### Testing Email Templates

```go
func TestEmailTemplate(t *testing.T) {
    app := setupTestApp() // Your test app setup
    
    variables := map[string]any{
        "AppName": "Test App",
        "Name":    "Test User",
        "Email":   "test@example.com",
        "Year":    "2025",
        "AppURL":  "http://localhost:8090",
    }
    
    err := SendCustomEmail(app, "test@example.com", "Test Email", "welcome", variables)
    if err != nil {
        t.Fatalf("Failed to send test email: %v", err)
    }
    
    // Check that job was queued
    // Verify email content in MailHog
}
```

## Troubleshooting

### Common Issues

1. **Template Not Found**
   - Ensure both `.html` and `.txt` files exist in `templates/emails/`
   - Check template name matches exactly (case-sensitive)

2. **SMTP Connection Failed**
   - Verify SMTP settings in `.env`
   - Check firewall and network connectivity
   - For Gmail, use app passwords instead of regular passwords

3. **Template Variables Not Rendering**
   - Ensure variable names match exactly in template and payload
   - Check Go template syntax (use `{{.VariableName}}`)

4. **Job Not Processing**
   - Verify job queue cron is enabled: `ENABLE_SYSTEM_QUEUE_CRON=true`
   - Check job queue worker configuration
   - Review application logs for processing errors

### Debugging

1. **Enable Debug Logging**:
   ```go
   logger := logger.GetLogger(app)
   logger.Debug("Email job details", "payload", payload)
   ```

2. **Check Job Queue Status**:
   ```bash
   # View queued jobs
   curl http://localhost:8090/api/collections/queues/records
   ```

3. **Monitor Email Sending**:
   - Check MailHog interface in development
   - Review SMTP server logs in production
   - Monitor application metrics for email success/failure rates

## Best Practices

1. **Template Organization**: Keep templates organized by purpose
2. **Variable Validation**: Validate required variables before sending
3. **Error Handling**: Always handle email sending errors gracefully
4. **Testing**: Test templates with various data scenarios
5. **Performance**: Use job queue for bulk emails to avoid blocking requests
6. **Security**: Sanitize user input in email content
7. **Monitoring**: Track email delivery rates and failures

## Integration Examples

### Custom API Endpoint

```go
// Custom route for sending notifications
func SendNotificationEmail(c echo.Context) error {
    var req struct {
        To       string         `json:"to"`
        Subject  string         `json:"subject"`
        Template string         `json:"template"`
        Data     map[string]any `json:"data"`
    }
    
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, map[string]string{"error": "Invalid request"})
    }
    
    app := c.Get("app").(*pocketbase.PocketBase)
    
    if err := SendCustomEmail(app, req.To, req.Subject, req.Template, req.Data); err != nil {
        return c.JSON(500, map[string]string{"error": "Failed to queue email"})
    }
    
    return c.JSON(200, map[string]string{"message": "Email queued successfully"})
}
```

This guide provides everything you need to implement custom emails in your PocketBase application using the job queue system with template support.