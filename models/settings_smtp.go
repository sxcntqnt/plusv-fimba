package models

import (
	"time"
)

type SettingsSMTP struct {
	ID		uint  `json:"smtp_id"`
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
