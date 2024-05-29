package models

import (
	"time"
)

type EmailTpl struct {
	Id         uint      `json:"et_id"`
	Et_name    string    `json:"et_name"`
	Et_subject string    `json:"et_subject"`
	Et_body    string    `json:"et_body"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP(6)"`
}
