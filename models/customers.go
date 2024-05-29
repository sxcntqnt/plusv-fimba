package models

import "time"

type Customer struct {
	Id            uint      `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email" "gorm:unique"`
	Address       string    `json:"address"`
	CreatedAt     time.Time `gorm:"type:TIMESTAMP(6)"`
	IsActive      string    `json:"isactive"`
	Modified_date time.Time `gorm:"type:TIMESTAMP(6)"`
	Password      []byte    `"json:"-"`
}
