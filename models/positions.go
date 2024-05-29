package models

import (
	"time"
)

type Positions struct {
	ClientId       uint      `json:"id"`
	ClientTime     time.Time `gorm:"type:TIMESTAMP(6)"`
	VehicleId      uint      `json:"v_id"`
	ClientLat      float64   `json:"latitude"`
	ClientLng      float64   `json:"longitude"`
	ClientStatus   string    `json:"Raining/Clear"`
	ClientAltitude float64   `json:"altitude"`
	ClientSpeed    float64   `json:"speed"`
	ClientBearing  float64   `json:"bearing"`
	ClientAccuracy int       `json:"accuracy"`
	ClientProvider string    `json:"provider"`
	ClientComment  string    `json:"comment"`
	CreatedAt      time.Time `gorm:"type:TIMESTAMP(6)"`
}
