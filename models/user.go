package models

import (
	"time"
)

type User struct {
	UId         uint   `json:"uid"`
	UName       string `json:"uname"`
	UUsername   string `json:"uusername"`
	UEmail      string `json:"email" gorm:"unique"`
	UPassword   string `json:"-"` // the dash indicates that this field should not be JSON-encoded
	URole       string `json:"role"`
	UIsActive   bool   `json:"active"`
	Permissions *Login_roles
	LoginData       *LoginData // new field for login information
	UCreatedAt  time.Time
}
