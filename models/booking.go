package models

import (
	"time"
)

type Booking struct {
	TripID                uint      `json:"t_id"`
	TripCustomerId        uint      `json:"t_customer_id"`
	TripVehicle           string    `json:"t_vechicle"`
	TripType              string    `json:"t_type"`
	TripDriver            string    `json:"t_driver"`
	TripStartDate         time.Time `gorm:"type:timestamp(6)"`
	TripEndDate           time.Time `gorm:"type:timestamp(6)"`
	TripSquadFrmLoc       string    `json:"t_trip_fromlocation"`
	TripSquadToLoc        string    `json:"t_trip_tolocation"`
	TripSquadFrmLat       string    `json:"t_trip_fromlat"`
	TripSquadFrmLng       string    `json:"t_trip_fromlog"`
	TripSquadToLat        string    `json:"t_trip_tolat"`
	TripSquadToLng        string    `json:"t_trip_tolog"`
	TripSquadTotalDist    string    `json:"t_totaldistance"`
	TripSquadAmount       string    `json:"t_trip_amount"`
	TripSquadStatus       string    `json:"t_trip_status"`
	TripSquadTrackingCode string    `json:"t_trackingcode"`
	TripSquadCreated_by   string    `json:"createdBy"`
	CreatedAt             time.Time `gorm:"type:TIMESTAMP(6)"`
	UpdatedAt             time.Time `gorm:"type:TIMESTAMP(6)"`
}
