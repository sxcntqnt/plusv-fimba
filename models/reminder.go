package models

import (
	"time"
)

type Reminder struct {
	RmdId     uint      `json:"r_id"`
	RmdDate   time.Time `gorm"timestamp(6)"`
	RmdMsg    string    `json:"r_message"`
	RmdIsRead string    `json:"r_isread"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6)"`
	Vehicle   Vehicle   `gorm:"foreignKey:VehicleID"`
}
