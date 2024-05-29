package routes

import (
	"context"
	"fmt"
	"log"

//	"fimba/database"
	"fimba/models"

	"github.com/gofiber/fiber/v2"
	mailslurp "github.com/mailslurp/mailslurp-client-go"
)

var smtpSettings models.SettingsSMTP
var mailslurpClient *mailslurp.APIClient
var mailslurpInbox *mailslurp.InboxDto

// SMTPConfigTestEmail sends a test email to confirm SMTP settings are working.
func SMTPConfigTestEmail(c *fiber.Ctx, apiKey string, client *mailslurp.APIClient, inbox *mailslurp.InboxDto, smtpSettings models.SettingsSMTP) error {
        // Parse JSON request body
       /* var smtpSettings {
                ReplyTo string `json:"smtp_replyto"`
        }
        if err := c.BodyParser(&smtpSettings); err != nil {
                return err
        }
*/
	replyTo := smtpSettings.SmtpReplyTo

        // Check if SMTP settings exist
        if smtpSettings.ID == 0 {
                log.Printf("SMTP settings do not exist")
                return fmt.Errorf("SMTP settings do not exist")
        }

        // Generate new email address for inbox
        ctx := context.Background()
        createInboxOpts := &mailslurp.CreateInboxOpts{}
        emailAddress, _, err := client.InboxControllerApi.CreateInbox(ctx, createInboxOpts)
        if err != nil {
                log.Printf("Failed to generate email address: %s", err)
                return fmt.Errorf("failed to generate email address")
        }

        // Send email
        subject := "SMTP Config Test"
        body := "This is a test email to confirm SMTP settings are working."
        sendEmailOptions := mailslurp.SendEmailOptions{
                To:      &[]string{replyTo},
                From:    &emailAddress.EmailAddress,
                ReplyTo: &replyTo,
                Subject: &subject,
                Body:    &body,
        }
        _, err = client.InboxControllerApi.SendEmail(ctx, inbox.Id, sendEmailOptions)
        if err != nil {
                log.Printf("Failed to send email: %s", err)
                return fmt.Errorf("failed to send email")
        }

        // Return success response
        return c.JSON(map[string]interface{}{
                "message": "Email sent successfully",
        })
}

// Cleanup deletes the MailSlurp inbox.
func Cleanup(ctx context.Context, client *mailslurp.APIClient, inbox *mailslurp.InboxDto) {
	_, err := client.InboxControllerApi.DeleteInbox(ctx, inbox.Id)
	if err != nil {
		log.Printf("MailSlurp inbox deletion failed: %s", err)
	}
}

func InitMailSlurpClient(apiKey string) (*mailslurp.InboxDto, *mailslurp.APIClient, error) {
	// Create a context with your api key
	ctx := context.WithValue(context.Background(), mailslurp.ContextAPIKey, mailslurp.APIKey{Key: apiKey})

	// Create MailSlurp client
	cfg := mailslurp.NewConfiguration()
	client := mailslurp.NewAPIClient(cfg)

	// Create inbox
	inboxOpts := &mailslurp.CreateInboxOpts{}
	mailslurpInbox, _, err := client.InboxControllerApi.CreateInbox(ctx, inboxOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create MailSlurp inbox: %v", err)
	}

	return &mailslurpInbox, client, nil
}

// Define a function that takes the required arguments and returns the handler function
func SmtpConfigEmailHandler(apiKey string, client *mailslurp.APIClient, inbox *mailslurp.InboxDto) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Call the original SMTPConfigTestEmail function with the provided arguments and the request context
		err := SMTPConfigTestEmail(c, apiKey, client, inbox, smtpSettings)
		if err != nil {
			// Handle errors
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return nil
	}
}
