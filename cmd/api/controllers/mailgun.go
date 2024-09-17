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

func SendEmail(domain, apiKey, emailTo, userName, userID, emailType string, delay int) (string, error) {
	var htmlFilename string
	var emailSubject string
	var msg string

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
	// mg.SetAPIBase(mailgun.APIBaseEU)

	htmlContent, err := os.ReadFile(fmt.Sprintf("template/%s.html", htmlFilename))
	if err != nil {
		return "", err
	}

	m := mg.NewMessage(
		fmt.Sprintf("Authentication API <mailgun@%s>", domain),
		fmt.Sprintf("%s %s", emailSubject, userName),
		msg,
		emailTo,
	)

	// Set the HTML content of the message using a template
	tmpl, err := template.New("email").Parse(string(htmlContent))
	if err != nil {
		return "", err
	}

	// Execute the template with the provided data
	err = tmpl.Execute(os.Stdout, user)
	if err != nil {
		panic(err)
	}

	// Execute the template with the provided data
	var emailBodyBuffer bytes.Buffer
	err = tmpl.Execute(&emailBodyBuffer, user)
	if err != nil {
		return "", err
	}

	// Set the HTML content of the message
	m.SetHtml(emailBodyBuffer.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Wait for the specified delay before sending the email
	time.Sleep(time.Duration(delay) * time.Second)

	// Sending the email
	_, id, err := mg.Send(ctx, m)
	return id, err
}

func ContactSubmit(domain, apiKey, userEmail, msg string) (string, error) {
	mg := mailgun.NewMailgun(domain, apiKey)
	m := mg.NewMessage(
		fmt.Sprintf("Authentication API <mailgun@%s>", domain),
		fmt.Sprintf("Contact form submitted from %s", userEmail),
		msg,
		"flabelatus@gmail.com",
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := mg.Send(ctx, m)
	return id, err
}
