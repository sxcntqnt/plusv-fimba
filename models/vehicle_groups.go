package models

import (
	"time"
)

type VehicleGroups struct {
	GroupID   uint      `json:"gr_id"`
	GroupName string    `json:"gr_name"`
	GroupDesc string    `json:"gr_desc"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6)"`

	// Define a one-to-many relationship between VehicleGroup and Vehicle
	Vehicles []Vehicle `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
