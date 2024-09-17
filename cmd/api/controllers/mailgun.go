// Package controllers provides core business logic for managing user operations,
// including password hashing, sending emails, and other related functions.
// It serves as the intermediate layer between the repository and handler layers,
// handling the logic that coordinates user data and processes.
package controllers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/mailgun/mailgun-go/v3"
)

type UserData struct {
	UserID   string
	Username string
}

// SendEmail sends an email using the Mailgun API. It supports three types of emails: password reset, password change, and account confirmation.
// The email content is generated using HTML templates and personalized with the user's data.
//
// Parameters:
// - domain: The Mailgun domain used to send the email.
// - apiKey: The Mailgun API key.
// - emailTo: The recipient's email address.
// - userName: The recipient's username (for personalization).
// - userID: The recipient's user ID (for personalization and link generation).
// - emailType: The type of email to send ("password", "passwordChange", or "confirmation").
// - delay: The delay (in seconds) before sending the email.
//
// Returns:
// - string: The ID of the email sent by Mailgun (if successful).
// - error: An error if the email fails to send or any step in the process fails.
func SendEmail(domain, apiKey, emailTo, userName, userID, emailType string, delay int) (string, error) {
	var htmlFilename string
	var emailSubject string
	var msg string

	// Determine the email template and subject based on the emailType
	switch emailType {
	case "password":
		htmlFilename = "resetPassword"
		emailSubject = "Password Reset for"
		msg = "You have requested to reset your password, in order to continue please click on the following link"
	case "passwordChange":
		htmlFilename = "passwordChanged"
		emailSubject = "Password Was Successfully Updated"
		msg = "You have successfully updated your password"
	case "confirmation":
		htmlFilename = "confirmationEmail"
		emailSubject = "Sign Up Confirmation for"
		msg = "Thank you for registering at Authentication API! Please verify your email to confirm your account using this link"
	}

	user := UserData{UserID: userID, Username: userName}

	mg := mailgun.NewMailgun(domain, apiKey)

	// Read the HTML template file
	htmlContent, err := os.ReadFile(fmt.Sprintf("template/%s.html", htmlFilename))
	if err != nil {
		return "", err
	}

	// Create a new email message
	m := mg.NewMessage(
		fmt.Sprintf("Authentication API <mailgun@%s>", domain),
		fmt.Sprintf("%s %s", emailSubject, userName),
		msg,
		emailTo,
	)

	// Parse and execute the HTML template with user data
	tmpl, err := template.New("email").Parse(string(htmlContent))
	if err != nil {
		return "", err
	}

	var emailBodyBuffer bytes.Buffer
	err = tmpl.Execute(&emailBodyBuffer, user)
	if err != nil {
		return "", err
	}

	// Set the HTML content of the message
	m.SetHtml(emailBodyBuffer.String())

	// Set up the context and send the email after a delay
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	time.Sleep(time.Duration(delay) * time.Second)

	// Send the email
	_, id, err := mg.Send(ctx, m)
	return id, err
}

// ContactSubmit sends a contact form submission via email using the Mailgun API.
// The message is sent to the application's support email address.
//
// Parameters:
// - domain: The Mailgun domain used to send the email.
// - apiKey: The Mailgun API key.
// - userEmail: The email address of the user submitting the form.
// - msg: The message body submitted by the user.
//
// Returns:
// - string: The ID of the email sent by Mailgun (if successful).
// - error: An error if the email fails to send or any step in the process fails.
func ContactSubmit(domain, apiKey, userEmail, msg string) (string, error) {
	mg := mailgun.NewMailgun(domain, apiKey)

	// Create a new email message for the contact form submission
	m := mg.NewMessage(
		fmt.Sprintf("Authentication API <mailgun@%s>", domain),
		fmt.Sprintf("Contact form submitted from %s", userEmail),
		msg,
		"flabelatus@gmail.com", // Destination email (support address)
	)

	// Set up the context and send the email
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Send the email
	_, id, err := mg.Send(ctx, m)
	return id, err
}
