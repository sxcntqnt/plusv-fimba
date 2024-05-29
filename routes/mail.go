package routes

import (
	"fimba/database"
	"fimba/models"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

type Settings_smtp struct {
        ID              uint  `json:"smtp_id"`
	SmtpHost      string    `json:"smtp_host"`
	SmtpAuth      string    `json:"smtp_auth"`
	SmtpUname     string    `json:"smtp_uname"`
	SmtpPwd       string    `json:"smtp_pwd"`
	SmtpIsSecure  string    `json:"smtp_issecure"`
	SmtpPort      string    `json:"smtp_port"`
	SmtpEmailFrom string    `json:"smtp_emailfrom"`
	SmtpReplyTo   string    `json:"smtp_replyto"`
	CreatedAt     time.Time `gorm:"type:TIMESTAMP(6)"`
}

type SMTPConfig struct {
	Host      string
	Port      string
	Username  string
	Password  string
	IsSecure  string
	EmailFrom string
	ReplyTo   string
}

func NewSMTPConfig(settings *Settings_smtp) (*SMTPConfig, error) {
	// Set up encryption method
	var IsSecure string
	if settings.SmtpIsSecure == "SSL" {
		IsSecure = "ssl"
	} else if settings.SmtpIsSecure == "TLS" {
		IsSecure = "tls"
	} else {
		IsSecure = "true" // set IsSecure to "true" by default
	}

	// Create the SMTPConfig struct
	config := &SMTPConfig{
		Host:      settings.SmtpHost,
		Port:      settings.SmtpPort,
		Username:  settings.SmtpUname,
		Password:  settings.SmtpPwd,
		IsSecure:  IsSecure,
		EmailFrom: settings.SmtpEmailFrom,
		ReplyTo:   settings.SmtpReplyTo,
	}

	return config, nil
}

func Smtpconfigsave(c *fiber.Ctx) error {
	// Get SMTP settings from configuration file
 	viper.SetConfigFile("./config/settings.json")
        	if err := viper.ReadInConfig(); err != nil {
                	return err
        	}
	smtpSettings := viper.GetStringMap("smtp")

	// Create new Settings_smtp instance with the SMTP configuration data
	s := Settings_smtp{
		SmtpHost:      smtpSettings["smtp_host"].(string),
		SmtpAuth:      smtpSettings["smtp_auth"].(string),
		SmtpUname:     smtpSettings["smtp_uname"].(string),
		SmtpPwd:       smtpSettings["smtp_pwd"].(string),
		SmtpIsSecure:  smtpSettings["smtp_issecure"].(string),
		SmtpPort:      smtpSettings["smtp_port"].(string),
		SmtpEmailFrom: smtpSettings["smtp_emailfrom"].(string),
		SmtpReplyTo:   smtpSettings["smtp_replyto"].(string),
		CreatedAt:     time.Now(),
	}

	// Check if SMTP configuration already exists in database
        var count int64
        if err := database.Database.Db.Model(&Settings_smtp{}).Count(&count).Error; err != nil {
                return err
        }
	if count == 0 {
		// Create new SMTP configuration in the database
		if err := database.Database.Db.Create(&s).Error; err != nil {
			return err
		}
	} else {
		// Update SMTP configuration in the database
		if err := database.Database.Db.Save(&s).Error; err != nil {
			return err
		}
	}

	// Return success response
	return c.JSON(s)
}

func EmailTemplate(c *fiber.Ctx) error {
	// Parse ID parameter from URL
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid template ID")
	}

	// Load email template from database
	var tpl models.EmailTpl
	result := database.Database.Db.First(&tpl, id)
	if result.Error != nil {
		return fiber.NewError(fiber.StatusNotFound, "Template not found")
	}

	// Create response payload
	payload := map[string]interface{}{
		"id":         tpl.Id,
		"et_name":    tpl.Et_name,
		"et_subject": tpl.Et_subject,
		"et_body":    tpl.Et_body,
		"CreatedAt":  tpl.CreatedAt,
	}

	// Return response
	return c.JSON(payload)
}
func UpdateEmailTemplate(c *fiber.Ctx) error {
	// Get template ID from URL parameter
	id := c.Params("id")

	// Parse JSON request body
	var template models.EmailTpl
	if err := c.BodyParser(&template); err != nil {
		return err
	}

	// Find template in the database
	result := database.Database.Db.First(&models.EmailTpl{}, id)
	if result.Error != nil {
		return result.Error
	}

	// Update template fields
	result = database.Database.Db.Model(&template).Updates(models.EmailTpl{
		Et_name:    template.Et_name,
		Et_subject: template.Et_subject,
		Et_body:    template.Et_body,
	})
	if result.Error != nil {
		return result.Error
	}

	// Return success response
	return c.JSON(template)
}
func SendMail(settings *Settings_smtp, to []string, subject string, body string) error {
	// Create the SMTP client config
	smtpConfig, err := NewSMTPConfig(settings)
	if err != nil {
		return err
	}

	// Set up the message headers
	headers := make(map[string]string)
	headers["From"] = smtpConfig.EmailFrom
	headers["To"] = strings.Join(to, ", ")
	headers["Subject"] = subject

	// Set up the message body
	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + body

	// Set up the SMTP server address
	serverAddr := fmt.Sprintf("%s:%s", smtpConfig.Host, smtpConfig.Port)

	// Set up authentication mechanism
	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)

	// Send the email
	err = smtp.SendMail(serverAddr, auth, smtpConfig.EmailFrom, to, []byte(message))
	if err != nil {
		return err
	}

	return nil
}
