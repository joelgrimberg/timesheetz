package email

import (
	"fmt"
	"os"
	"timesheet/internal/config"

	"github.com/resend/resend-go/v2"
)

func EmailAttachment(filename string) {
	// Get email configuration from config
	name, sendToOthers, recipientEmail, senderEmail, replyToEmail, apiKey, err := config.GetEmailConfig()
	if err != nil {
		fmt.Println("Error loading email configuration:", err.Error())
		return
	}
	// Check if user wants to send EmailAttachment
	if !sendToOthers {
		fmt.Println("not sending to others")
	}

	client := resend.NewClient(apiKey)

	// Read attachment file
	pwd, _ := os.Getwd()
	f, err := os.ReadFile(pwd + "/" + filename)
	if err != nil {
		fmt.Println("Error reading attachment file:", err.Error())
		return
	}

	// Create attachments objects
	pdfAttachmentFromLocalFile := &resend.Attachment{
		Content:     f,
		Filename:    filename,
		ContentType: "application/image",
	}

	// Set up recipients
	recipients := []string{recipientEmail}
	if sendToOthers {
		// Add additional recipients if configured to send to others
		// You might want to read these from config as well
	}

	// Prepare email parameters
	params := &resend.SendEmailRequest{
		From:        name + "<" + senderEmail + ">",
		To:          recipients,
		Html:        "<strong>Timesheetz brought to you by a unicorn</strong>",
		Subject:     "urensheet " + name,
		Cc:          []string{},
		Bcc:         []string{},
		ReplyTo:     replyToEmail,
		Attachments: []*resend.Attachment{pdfAttachmentFromLocalFile},
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		fmt.Println("Error sending email:", err.Error())
		return
	}
	fmt.Println("Email sent successfully, ID:", sent.Id)
}
