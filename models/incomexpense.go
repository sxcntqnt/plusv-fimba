package models

import (
	"time"
)

// add foreign id from vehicle table
type Incomexpe struct {
	IeID          uint      `json:"ie_id"`
	IeVehicleID   string    `json:"ie_v_id"`
	IeDate        time.Time `json:"ie_date"`
	IeType        string    `json:"ie_type"`
	IeDescription string    `json:"ie_description"`
	IeAmount      int       `json:"ie_amount"`
	IsIncome      bool      `json:"ie_income"`
	IsExpense     bool      `json:"is_expense"`

	CreatedAt time.Time `gorm:"type:TIMESTAMP(6)"`
	UpdatedAt time.Time `gorm:"type:TIMESTAMP(6)"`
}

