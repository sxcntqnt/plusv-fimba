package models

import (
	"time"
)

// add foreign key from table vehicle
type FuelEntry struct {
	Id                uint      `json:"v_fuel_id"`
	VehicleID         int       `json:"v_id"`
	V_fuel_quantity   float64   `json:"v_fuel_quantity"`
	V_odometerreading int       `json:"v_odometerreading"`
	V_fuelprice       float64   `json:"v_fuelprice"`
	V_fuelfilldate    time.Time `gorm:"type:TIMESTAMP(6)"`
	V_fueladdedby     string    `json:"v_fueladdedby"`
	V_fuelcomments    string    `json:"v_fuelcomments"`
	CreatedAt         time.Time `gorm:"type:TIMESTAMP(6)"`
	Vehicle           Vehicle   `gorm:"foreignKey:VehicleID"`
}
