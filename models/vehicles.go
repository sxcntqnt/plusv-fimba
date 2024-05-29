package models

import "time"

type Vehicle struct {
	V_Id              uint      `json:"v_id" gorm:"primaryKey"`
	V_RegistrationNo  string    `json:"v_registration_no"`
	V_Name            string    `json:"v_name"`
	V_Model           string    `json:"v_model"`
	V_ChassisNo       string    `json:"v_chassis_no"`
	V_EngineNo        string    `json:"v_engine_no"`
	V_Manufactured_by string    `json:"v_manufactured_by"`
	V_Type            string    `json:"v_type"`
	V_Color           string    `json:"v_color"`
	V_Mileageperlitre string    `json:"v_mileageperlitre"`
	V_Is_active       int       `json:"v_is_active"`
	V_Group           int       `json:"v_group"`
	V_Reg_exp_date    string    `json:"v_reg_exp_date"`
	V_Api_url         string    `json:"v_api_url"`
	V_Api_username    string    `json:"v_api_username"`
	V_Api_password    string    `json:"v_api_password"`
	V_Created_by      string    `json:"v_created_by"`
	V_CreatedAt       time.Time `gorm:"type:timestamp(6)"`
	V_modified_date   time.Time `gorm:"type:timestamp(6)"`

	FuelEntries []FuelEntry `gorm:"foreignKey:VehicleID"`
	Reminders   []Reminder  `gorm:"foreignKey:VehicleID"`
}
