package models

import (
	"time"
)

type TripPayments struct {
	TPId        uint      `json:"tp_id"`
	TPTripId    int       `json:"tp_trip_id"`
	TPVehicleId int       `json:"tp_v_id"`
	TPAmount    int       `json:"tp_amount"`
	TPNotes     string    `json:"tp_notes"`
	CreatedAt   time.Time `gorm:"type:TIMESTAMP(6)"`
}
