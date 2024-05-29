package models

type Login_roles struct {
	Lr_id                uint `json:"lr_id"`
	LrUserId             int  `json:"lr_u_id"`
	LrVehicleList        int  `json:"lr_vech_list"`
	LrVehicleListView    int  `json:"lr_vech_list_view"`
	LrVehicleListEdit    int  `json:"lr_vech_list_edit"`
	LrVehicleAdd         int  `json:"lr_vech_add"`
	LrVehicleGroup       int  `json:"lr_vech_group"`
	LrVehicleGroupAdd    int  `json:"lr_vech_group_add"`
	LrVehicleGroupAction int  `json:"lr_vech_group_action"`
	LrDriversList        int  `json:"lr_drivers_list"`
	LrDriversListEdit    int  `json:"lr_drivers_list_edit"`
	LrDriversAdd         int  `json:"lr_drivers_add"`
	LrTripsList          int  `json:"lr_trips_list"`
	LrTripsListEdit      int  `json:"lr_trips_list_edit"`
	LrTripsAdd           int  `json:"lr_trips_add"`
	LrCustomerList       int  `json:"lr_cust_list"`
	LrCustomerEdit       int  `json:"lr_cust_edit"`
	LrCustomerAdd        int  `json:"lr_cust_add"`
	LrFuelList           int  `json:"lr_fuel_list"`
	LrFuelEdit           int  `json:"lr_fuel_edit"`
	LrFuelAdd            int  `json:"lr_fuel_add"`
	LrReminderList       int  `json:"lr_reminder_list"`
	LrReminderDelete     int  `json:"lr_reminder_delete"`
	LrReminderAdd        int  `json:"lr_reminder_add"`
	LrIeList             int  `json:"lr_ie_list"`
	LrIeEdit             int  `json:"lr_ie_edit"`
	LrIeAdd              int  `json:"lr_ie_add"`
	LrIeTracking         int  `json:"lr_tracking"`
	LrLiveLoc            int  `json:"lr_liveloc"`
	LrGeofenceAdd        int  `json:"lr_geofence_add"`
	LrGeofenceList       int  `json:"lr_geofence_list"`
	LrGeofenceDelete     int  `json:"lr_geofence_delete"`
	LrGeofenceEvents     int  `json:"lr_geofence_events"`
	LrReports            int  `json:"lr_reports"`
	LrSettings           int  `json:"lr_settings"`
}
