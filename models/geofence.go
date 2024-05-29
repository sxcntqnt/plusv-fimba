package models

import (
	"time"
)

type Geofence struct {
	Geo_id          uint      `json:"geo_id"`
	Geo_name        string    `json:"geo_name"`
	Geo_description string    `json:"geo_description"`
	Geo_area        string    `json:"geo_area"`
	Geo_vehicles    string    `json:"geo_vehicles"`
	CreatedAt       time.Time `		gorm:"type:TIMESTAMP(6)"`
	UpdatedAt       time.Time `gorm:"type:TIMESTAMP(6)"`
}
