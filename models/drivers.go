package models

import "time"

type Driver struct {
	DID              uint      `json:"d_id"`
	DName            string    `json:"d_name"`
	DMobile          string    `json:"d_mobile"`
	DAddress         string    `json:"d_address"`
	DAge             int       `json:"d_age"`
	DLicenseNo       string    `json:"d_licenseno"`
	DLicenceExpdate  time.Time `gorm:"type:TIMESTAMP(6)"`
	DTotalExperience int       `json:"d_total_exp"`
	DDateOfJoining   time.Time `gorm:"type:TIMESTAMP(6)"`
	DReference       string    `json:"d_ref"`
	DIsActive        int       `json:"d_is_active"`
	DCreatedBy       uint      `json:"d_created_by"`
	DCreatedAt       time.Time `gorm:"type:TIMESTAMP(6)"`
	DModifiedDate    time.Time `gorm:"type:TIMESTAMP(6)"`
}
