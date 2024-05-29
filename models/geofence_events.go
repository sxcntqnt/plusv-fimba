package models

import (
	"time"
)

// add foreign id from vehicle table and geofence
type Geofence_events struct {
	ID           uint      `gorm:"primary_key" json:"id"`
	GeVehicleId  string    `json:"ge_v_id"`
	GeGeofenceId string    `json:"ge_geo_id"`
	Geevent      string    `json:"ge_event"`
	Getimestamp  string    `json:"type:timestamp(6)"`
	CreatedAt    time.Time `gorm:"type:TIMESTAMP(6)"`
}
