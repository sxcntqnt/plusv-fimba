package routes

import (
	"fimba/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Dashboard struct {
	db *gorm.DB
}

type IeChartData struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}
type DashReminder struct {
	RmdId        uint      `json:"r_id"`
	RmdVehicleId int       `json:"r_v_id"`
	RmdDate      time.Time `gorm"timestamp(6)"`
	RmdMsg       string    `json:"r_message"`
	RmdIsRead    string    `json:"r_isread"`
}

func NewDashboard(db *gorm.DB) *Dashboard {
	return &Dashboard{db}
}

type DashboardInfo struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type VehicleLocation struct {
	VehicleID uint    `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type VehicleStatus struct {
	VehicleID uint   `json:"vehicle_id"`
	Status    string `json:"status"`
}

type GeofenceEvent struct {
	ID           uint      `json:"id"`
	GeoFenceID   uint      `json:"ge_geo_id"`
	VehicleID    uint      `json:"vehicle_id"`
	EnteredAt    time.Time `json:"entered_at"`
	ExitedAt     time.Time `json:"exited_at"`
	GeoFenceName string    `json:"geo_name"`
}

func DashboardHandler(c *fiber.Ctx) error {
	d := new(Dashboard)

	chartData, _ := d.GetIeChartData()
	infoData := d.GetDashboardInfo(c)
	reminderData, _ := d.GetTodayReminder()
	statusData, _ := d.GetVehicleStatus()

	data := fiber.Map{
		"chartData":    chartData,
		"infoData":     infoData,
		"reminderData": reminderData,
		"statusData":   statusData,
	}
	var returndata []map[string]interface{}
	geofenceevents, _ := d.GetGeofenceEvents(20)
	if len(geofenceevents) > 0 {
		for _, eventData := range geofenceevents {
			geoID, _ := strconv.ParseUint(eventData.GeGeofenceId, 10, 64)
			geoName, _ := d.GetGeofenceName(uint(geoID))
			if geoName != "" {
				m := make(map[string]interface{})
				m["ge_vehicle_id"] = eventData.GeVehicleId
				m["ge_geofence_id"] = eventData.GeGeofenceId
				m["ge_event"] = eventData.Geevent
				m["ge_timestamp"] = eventData.Getimestamp
				m["geofence_name"] = geoName
				returndata = append(returndata, m)
			}
		}
	}
	data["geofenceevents"] = returndata

	return c.JSON(data)
}

func (d *Dashboard) GetGeofenceName(id uint) (string, error) {
	var geofence models.Geofence
	if err := d.db.Where("geo_id = ?", id).First(&geofence).Error; err != nil {
		return "", err
	}
	return geofence.Geo_name, nil
}

func (d *Dashboard) GetGeofenceEvents(limit int) ([]models.Geofence_events, error) {
	var geofenceEvents []models.Geofence_events
	err := d.db.Limit(limit).Find(&geofenceEvents).Error
	if err != nil {
		return nil, err
	}
	return geofenceEvents, nil
}
func (d *Dashboard) GetVehicleStatus() ([]models.Positions, error) {
	var vehicleStatus []models.Positions
	if err := d.db.Find(&vehicleStatus).Error; err != nil {
		return nil, err
	}
	return vehicleStatus, nil
}

func (d *Dashboard) GetIeChartData() (map[string]IeChartData, error) {
	var data []struct {
		Date    string
		Income  float64
		Expense float64
	}

	err := d.db.Table("incomexpe").Select("date, SUM(ie_amount) as income, SUM(CASE ie_type WHEN 'expense' THEN ie_amount ELSE 0 END) as expense").Group("date").Scan(&data).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]IeChartData)
	for _, row := range data {
		result[row.Date] = IeChartData{row.Income, row.Expense}
	}

	return result, nil
}

func (d *Dashboard) GetTodayReminder() ([]models.Reminder, error) {
	var reminders []models.Reminder
	err := d.db.Where("r_date = ? AND r_isread = ?", time.Now().Format("2006-01-02"), 0).Find(&reminders).Error
	if err != nil {
		return nil, err
	}
	return reminders, nil
}

func (d *Dashboard) GetDashboardInfo(c *fiber.Ctx) error {
	var dashboardInfo []DashboardInfo

	if err := d.db.Find(&dashboardInfo).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	data := make(map[string]interface{})
	for _, info := range dashboardInfo {
		data[info.Name] = info.Value
	}

	return c.JSON(data)
}
