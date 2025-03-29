package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// EmailService handles email operations
type EmailService struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	BaseURL      string
}

// EmailData contains the data to be used in email templates
type EmailData struct {
	Name            string
	VerificationURL string
	ResetURL        string
	Year            int
}

// Default templates as fallbacks
const defaultVerificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verifikasi Email DelPresence</title>
</head>
<body style="font-family: Arial, sans-serif; margin: 0; padding: 20px; color: #333; background-color: #f7f7f7;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <div style="background-color: #0687C9; padding: 20px; text-align: center;">
            <h1 style="color: #fff; margin: 0; font-size: 24px;">DelPresence</h1>
        </div>
        <div style="padding: 30px;">
            <h2 style="margin-top: 0; color: #333;">Verifikasi Email Anda</h2>
            <p>Halo {{.Name}},</p>
            <p>Terima kasih telah mendaftar di DelPresence. Silakan klik tombol di bawah untuk memverifikasi alamat email Anda:</p>
            <div style="text-align: center; margin: 30px 0;">
                <a href="{{.VerificationURL}}" style="background-color: #0687C9; color: #ffffff; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold; display: inline-block;">Verifikasi Email</a>
            </div>
            <p>Jika Anda tidak dapat mengklik tombol di atas, salin dan tempel URL berikut ke browser Anda:</p>
            <p style="word-break: break-all; background-color: #f5f5f5; padding: 10px; border-radius: 4px; font-size: 14px;">{{.VerificationURL}}</p>
            <p>Tautan ini akan kedaluwarsa dalam 24 jam.</p>
            <p>Jika Anda tidak membuat permintaan ini, Anda dapat mengabaikan email ini.</p>
        </div>
        <div style="background-color: #f5f5f5; padding: 20px; text-align: center; font-size: 12px; color: #666;">
            <p>&copy; {{.Year}} DelPresence - Institut Teknologi Del. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const defaultResetPasswordTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Password DelPresence</title>
</head>
<body style="font-family: Arial, sans-serif; margin: 0; padding: 20px; color: #333; background-color: #f7f7f7;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
        <div style="background-color: #0687C9; padding: 20px; text-align: center;">
            <h1 style="color: #fff; margin: 0; font-size: 24px;">DelPresence</h1>
        </div>
        <div style="padding: 30px;">
            <h2 style="margin-top: 0; color: #333;">Reset Password Anda</h2>
            <p>Halo {{.Name}},</p>
            <p>Kami menerima permintaan untuk mengatur ulang password akun DelPresence Anda. Silakan klik tombol di bawah untuk melanjutkan:</p>
            <div style="text-align: center; margin: 30px 0;">
                <a href="{{.ResetURL}}" style="background-color: #0687C9; color: #ffffff; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold; display: inline-block;">Reset Password</a>
            </div>
            <p>Jika Anda tidak dapat mengklik tombol di atas, salin dan tempel URL berikut ke browser Anda:</p>
            <p style="word-break: break-all; background-color: #f5f5f5; padding: 10px; border-radius: 4px; font-size: 14px;">{{.ResetURL}}</p>
            <p>Tautan ini akan kedaluwarsa dalam 1 jam.</p>
            <p>Jika Anda tidak membuat permintaan ini, Anda dapat mengabaikan email ini dan password Anda akan tetap sama.</p>
        </div>
        <div style="background-color: #f5f5f5; padding: 20px; text-align: center; font-size: 12px; color: #666;">
            <p>&copy; {{.Year}} DelPresence - Institut Teknologi Del. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

// NewEmailService creates a new instance of EmailService
func NewEmailService() *EmailService {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
		log.Println("Using environment variables from system")
	} else {
		log.Println("Successfully loaded .env file")
	}

	// Try to load email-specific config
	_ = godotenv.Load(".env.email")
	log.Println("Attempted to load .env.email file")

	// Read and log all email-related environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("FROM_EMAIL")
	fromName := os.Getenv("FROM_NAME")
	baseURL := os.Getenv("FRONTEND_URL")

	// Fallback to Mailtrap defaults if env variables are not set
	if smtpHost == "" {
		smtpHost = "sandbox.smtp.mailtrap.io"
		log.Println("SMTP_HOST not found, using default: sandbox.smtp.mailtrap.io")
	}

	if smtpUsername == "" {
		smtpUsername = "9271bf618007ff"
		log.Println("SMTP_USERNAME not found, using default Mailtrap testing credentials")
	}

	if smtpPassword == "" {
		smtpPassword = "f6649995eff467"
		log.Println("SMTP_PASSWORD not found, using default Mailtrap testing credentials")
	}

	if fromEmail == "" {
		fromEmail = "noreply@delpresence.com"
		log.Println("FROM_EMAIL not found, using default: noreply@delpresence.com")
	}

	if fromName == "" {
		fromName = "DelPresence"
		log.Println("FROM_NAME not found, using default: DelPresence")
	}

	if baseURL == "" {
		baseURL = "http://localhost:3000"
		log.Println("FRONTEND_URL not found, using default: http://localhost:3000")
	}

	log.Printf("Email Environment Variables (after defaults):")
	log.Printf("- SMTP_HOST: %s", smtpHost)
	log.Printf("- SMTP_USERNAME: %s", smtpUsername)
	log.Printf("- SMTP_PASSWORD: %s", maskPassword(smtpPassword))
	log.Printf("- FROM_EMAIL: %s", fromEmail)
	log.Printf("- FROM_NAME: %s", fromName)
	log.Printf("- FRONTEND_URL: %s", baseURL)

	// Parse SMTP port with default fallback
	smtpPort := 2525 // Default untuk Mailtrap
	if portStr := os.Getenv("SMTP_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			smtpPort = port
		} else {
			log.Printf("Warning: Invalid SMTP_PORT '%s', using default port 2525 for Mailtrap", portStr)
		}
	}
	log.Printf("- SMTP_PORT: %d", smtpPort)

	return &EmailService{
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUsername: smtpUsername,
		SMTPPassword: smtpPassword,
		FromEmail:    fromEmail,
		FromName:     fromName,
		BaseURL:      baseURL,
	}
}

// maskPassword hides most of the password for logging
func maskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}

	return password[:2] + "****" + password[len(password)-2:]
}

// SendVerificationEmail sends an email with verification link
func (s *EmailService) SendVerificationEmail(email, name, token string) error {
	verificationURL := s.BaseURL + "/verify-email?token=" + token

	// Set up template data
	data := EmailData{
		Name:            name,
		VerificationURL: verificationURL,
		Year:            time.Now().Year(),
	}

	// Define possible template paths
	fileName := "verification.html"
	possiblePaths := []string{
		"internal/templates/email/" + fileName,
		"delpresence-api/internal/templates/email/" + fileName,
		"../../internal/templates/email/" + fileName,
		"../internal/templates/email/" + fileName,
	}

	// Add current working directory based paths
	if rootDir, err := os.Getwd(); err == nil {
		possiblePaths = append(possiblePaths,
			filepath.Join(rootDir, "internal/templates/email", fileName),
			filepath.Join(rootDir, "delpresence-api/internal/templates/email", fileName),
			filepath.Join(rootDir, "../internal/templates/email", fileName),
			filepath.Join(rootDir, "../../internal/templates/email", fileName))
	}

	// Add executable directory based paths
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		possiblePaths = append(possiblePaths,
			filepath.Join(execDir, "internal/templates/email", fileName),
			filepath.Join(execDir, "../internal/templates/email", fileName),
			filepath.Join(execDir, "../../internal/templates/email", fileName),
			filepath.Join(execDir, "../../../internal/templates/email", fileName))
	}

	// Try each path until a template is successfully loaded
	var tmpl *template.Template
	var err error
	for _, path := range possiblePaths {
		tmpl, err = template.ParseFiles(path)
		if err == nil {
			// Successfully loaded template
			log.Printf("Successfully loaded verification template from: %s", path)
			break
		}
	}

	// If template wasn't found, use default inline template
	if err != nil {
		log.Printf("Warning: Could not load verification template file, using default template. Error: %v", err)
		log.Printf("Tried the following paths: %v", possiblePaths)

		// Parse the default template string
		tmpl, err = template.New("verification").Parse(defaultVerificationTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse default verification template: %w", err)
		}
	}

	// Render template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	// Send email
	return s.sendEmail(email, "Verifikasi Email DelPresence", body.String())
}

// SendPasswordResetEmail sends an email with password reset link
func (s *EmailService) SendPasswordResetEmail(email, name, token string, isAdmin bool) error {
	// Different paths for admin and regular users
	resetPath := "/reset-password"
	if isAdmin {
		resetPath = "/admin-reset-password"
	}

	resetURL := s.BaseURL + resetPath + "?token=" + token

	// Set up template data
	data := EmailData{
		Name:     name,
		ResetURL: resetURL,
		Year:     time.Now().Year(),
	}

	// Define possible template paths
	fileName := "reset_password.html"
	possiblePaths := []string{
		"internal/templates/email/" + fileName,
		"delpresence-api/internal/templates/email/" + fileName,
		"../../internal/templates/email/" + fileName,
		"../internal/templates/email/" + fileName,
	}

	// Add current working directory based paths
	if rootDir, err := os.Getwd(); err == nil {
		possiblePaths = append(possiblePaths,
			filepath.Join(rootDir, "internal/templates/email", fileName),
			filepath.Join(rootDir, "delpresence-api/internal/templates/email", fileName),
			filepath.Join(rootDir, "../internal/templates/email", fileName),
			filepath.Join(rootDir, "../../internal/templates/email", fileName))
	}

	// Add executable directory based paths
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		possiblePaths = append(possiblePaths,
			filepath.Join(execDir, "internal/templates/email", fileName),
			filepath.Join(execDir, "../internal/templates/email", fileName),
			filepath.Join(execDir, "../../internal/templates/email", fileName),
			filepath.Join(execDir, "../../../internal/templates/email", fileName))
	}

	// Try each path until a template is successfully loaded
	var tmpl *template.Template
	var err error
	for _, path := range possiblePaths {
		tmpl, err = template.ParseFiles(path)
		if err == nil {
			// Successfully loaded template
			log.Printf("Successfully loaded password reset template from: %s", path)
			break
		}
	}

	// If template wasn't found, use default inline template
	if err != nil {
		log.Printf("Warning: Could not load password reset template file, using default template. Error: %v", err)
		log.Printf("Tried the following paths: %v", possiblePaths)

		// Parse the default template string
		tmpl, err = template.New("reset_password").Parse(defaultResetPasswordTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse default password reset template: %w", err)
		}
	}

	// Render template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	// Send email
	return s.sendEmail(email, "Reset Password DelPresence", body.String())
}

// sendEmail handles the actual email sending
func (s *EmailService) sendEmail(to, subject, body string) error {
	// Log koneksi SMTP
	log.Printf("SMTP Configuration: Host=%s, Port=%d, Username=%s, FromEmail=%s, FromName=%s",
		s.SMTPHost, s.SMTPPort, s.SMTPUsername, s.FromEmail, s.FromName)

	if s.SMTPHost == "" || s.SMTPUsername == "" || s.SMTPPassword == "" {
		log.Println("Email configuration is incomplete. Skipping email sending.")
		// For development, log the email content
		log.Printf("Would have sent email to %s with subject: %s", to, subject)
		if len(body) > 500 {
			log.Printf("Email body (first 500 chars): %s", body[:500])
		} else {
			log.Printf("Email body: %s", body)
		}
		return nil
	}

	// Check if we're running in a development environment without SMTP access
	if os.Getenv("NO_SMTP") == "true" || os.Getenv("ENV") == "development" {
		log.Println("Running in development mode or NO_SMTP=true. Using file-based email delivery.")
		return s.saveEmailToFile(to, subject, body)
	}

	// Set up authentication information
	log.Printf("Setting up SMTP authentication for %s", s.SMTPHost)
	auth := smtp.PlainAuth("", s.SMTPUsername, s.SMTPPassword, s.SMTPHost)

	// Compose the email
	from := fmt.Sprintf("%s <%s>", s.FromName, s.FromEmail)
	if s.FromName == "" {
		from = s.FromEmail
	}

	// Format the email headers
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"%s\r\n"+
		"%s", from, to, subject, mime, body)

	// Send the email
	addr := fmt.Sprintf("%s:%d", s.SMTPHost, s.SMTPPort)
	log.Printf("Attempting to send email via SMTP server: %s", addr)

	// For Mailtrap specifically, print the inbox URL
	if strings.Contains(s.SMTPHost, "mailtrap.io") {
		log.Printf("Using Mailtrap - check your inbox at https://mailtrap.io/inboxes to see the email")
	}

	err := smtp.SendMail(addr, auth, s.FromEmail, []string{to}, []byte(message))
	if err != nil {
		log.Printf("Error sending email: %v", err)
		log.Println("Falling back to file-based email delivery...")
		return s.saveEmailToFile(to, subject, body)
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}

// saveEmailToFile saves the email content to a file for development use
func (s *EmailService) saveEmailToFile(to, subject, body string) error {
	// Create emails directory if it doesn't exist
	emailDir := "emails"
	if err := os.MkdirAll(emailDir, 0755); err != nil {
		log.Printf("Failed to create emails directory: %v", err)
		return err
	}

	// Create a filename based on recipient and timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Join(emailDir, fmt.Sprintf("%s-%s.html", strings.Replace(to, "@", "-at-", -1), timestamp))

	// Create email content with headers
	emailContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .email-container { border: 1px solid #ddd; padding: 20px; max-width: 800px; margin: 0 auto; }
        .email-header { background-color: #f5f5f5; padding: 10px; margin-bottom: 20px; }
        .email-info { margin-bottom: 20px; }
        .email-body { border-top: 1px solid #eee; padding-top: 20px; }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="email-header">
            <h2>%s</h2>
        </div>
        <div class="email-info">
            <p><strong>To:</strong> %s</p>
            <p><strong>From:</strong> %s &lt;%s&gt;</p>
            <p><strong>Date:</strong> %s</p>
        </div>
        <div class="email-body">
            %s
        </div>
    </div>
</body>
</html>`, subject, subject, to, s.FromName, s.FromEmail, time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700"), body)

	// Write to file
	err := os.WriteFile(filename, []byte(emailContent), 0644)
	if err != nil {
		log.Printf("Failed to write email to file: %v", err)
		return err
	}

	absPath, _ := filepath.Abs(filename)
	log.Printf("Email saved to file: %s", absPath)
	log.Printf("Open this file in your browser to view the email")

	return nil
}

// Helper function to get minimum of two values
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
