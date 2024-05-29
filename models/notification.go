package models

import (
	"time"
)

type notification struct {
	ntfId      uint      `json:"n_id"`
	ntfSubject string    `json:"n_subject"`
	ntfMessage string    `json:"n_message"`
	ntfIsRead  int       `json:"n_is_read"`
	CreatedAt  time.Time `gorm:"type:TIMESTAMP(6)"`
}
