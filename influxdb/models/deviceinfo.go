package models

type DeviceInfo struct { //contains details about the device
	DeviceID           string `json:"deviceId"`
	Location           string `json:"location"`
	LastWriteTimestamp int64  `json:"lastWriteTimestamp"`
}
