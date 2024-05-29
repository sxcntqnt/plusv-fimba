package routes

import (
	"fmt"
	"strconv"
	"time"

	"fimba/database"
	"fimba/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IncomeExpense struct {
	IeID          uint      `json:"ie_id"`
	IeVehicleID   string    `json:"ie_v_id"`
	IeDate        string    `json:"date"`
	IsIncome      bool      `json:"ie_income"`
	IeDescription string    `json:"ie_description"`
	IeAmount      int       `json:"ie_amount"`
	IsExpense     bool      `json:"is_expense"`
	IeType        string    `json:"income_type"`
	CreatedAt     time.Time `gorm:"type:TIMESTAMP(6)"`
	UpdatedAt     time.Time `gorm:"type:TIMESTAMP(6)"`
}

type TripPayments struct {
	TPId        uint      `json:"tp_id"`
	TPTripId    int       `json:"tp_trip_id"`
	TPVehicleId int       `json:"tp_v_id"`
	TPAmount    int       `json:"tp_amount"`
	CreatedAt   time.Time `gorm:"type:TIMESTAMP(6)"`
}

func GenerateFuelReport(c *fiber.Ctx) error {
	// Get the start and end date from the query parameters
	startDate, err := time.Parse("2006-01-02", c.Query("start_date"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start date format. Use YYYY-MM-DD.",
		})
	}

	endDate, err := time.Parse("2006-01-02", c.Query("end_date"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end date format. Use YYYY-MM-DD.",
		})
	}

	// Query the database for all fuel entries within the given time range
	var fuels []models.FuelEntry
	database.Database.Db.Where("v_fuelfilldate BETWEEN ? AND ?", startDate, endDate).Find(&fuels)

	// Calculate the total fuel consumption and cost
	totalFuel := 0.0
	totalCost := 0.0

	for _, f := range fuels {
		fuelQty := f.V_fuel_quantity
		fuelPrice := f.V_fuelprice

		totalFuel += fuelQty
		totalCost += fuelPrice * fuelQty
	}

	// Return the fuel report as a JSON response
	return c.JSON(fiber.Map{
		"start_date":   startDate.Format("2006-01-02"),
		"end_date":     endDate.Format("2006-01-02"),
		"total_fuel":   totalFuel,
		"total_cost":   totalCost,
		"fuel_entries": fuels,
	})
}

func CreateReport(vehicleID uint, db *gorm.DB) string {
	// Query the database for the vehicle's expenses and income
	ieList := []models.Incomexpe{}
	database.Database.Db.Where("vehicle_id = ?", vehicleID).Find(&ieList)

	// Calculate the total expenses and income
	totalExpenses := 0
	totalIncome := 0
	for _, ie := range ieList {
		if ie.IeType == "expense" {
			totalExpenses += ie.IeAmount
		} else if ie.IeType == "income" {
			totalIncome += ie.IeAmount
		}
	}

	// Create the report
	report := fmt.Sprintf("Vehicle %d Report\n", vehicleID)
	report += fmt.Sprintf("Total Expenses: $%d\n", totalExpenses)
	report += fmt.Sprintf("Total Income: $%d\n", totalIncome)
	report += fmt.Sprintf("Net Income: $%d\n", totalIncome-totalExpenses)

	// Return the report as a string
	return report
}

func generateTripsReport() ([]TripPayments, error) {
	db := database.Database.Db
	trips := []TripPayments{}
	result := db.Find(&trips)
	if result.Error != nil {
		return nil, result.Error
	}
	return trips, nil
}

func GenerateBookingReport(c *fiber.Ctx) error {
	// Get the start and end date from the query parameters
	startDate, err := time.Parse("2006-01-02", c.Query("start_date"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start date format. Use YYYY-MM-DD.",
		})
	}

	endDate, err := time.Parse("2006-01-02", c.Query("end_date"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end date format. Use YYYY-MM-DD.",
		})
	}

	// Get the vehicle from the query parameters
	vehicle := c.Query("vehicle")

	// Query the database for all bookings within the given time range and vehicle
	var bookings []models.Booking
	database.Database.Db.Where("trip_start_date BETWEEN ? AND ? AND trip_vehicle = ?", startDate, endDate, vehicle).Find(&bookings)

	// Create a slice to store the booking reports
	var reports []map[string]interface{}

	// Iterate over each booking and extract the required data points
	for _, b := range bookings {
		report := map[string]interface{}{
			"customer":    b.TripCustomerId,
			"vehicle":     b.TripVehicle,
			"type":        b.TripType,
			"driver":      b.TripDriver,
			"from":        b.TripSquadFrmLoc,
			"to":          b.TripSquadToLoc,
			"distance":    b.TripSquadTotalDist,
			"amount":      b.TripSquadAmount,
			"trip_status": b.TripSquadStatus,
			"created_by":  b.TripSquadCreated_by,
			"start_date":  b.TripStartDate.Format("2006-01-02"),
			"end_date":    b.TripEndDate.Format("2006-01-02"),
		}
		reports = append(reports, report)
	}

	// Return the booking report as a JSON response
	return c.JSON(fiber.Map{
		"start_date":      startDate.Format("2006-01-02"),
		"end_date":        endDate.Format("2006-01-02"),
		"vehicle":         vehicle,
		"booking_count":   len(bookings),
		"booking_reports": reports,
	})
}

func GenerateIncExpReport(c *fiber.Ctx) error {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	vehicleIDStr := c.Query("vehicle_id")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start date format. Use YYYY-MM-DD.",
		})
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end date format. Use YYYY-MM-DD.",
		})
	}

	vehicleID, err := strconv.Atoi(vehicleIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid vehicle ID.",
		})
	}

	incExpList, err := GetIncExpList(startDate, endDate, vehicleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get income/expense entries from database.",
		})
	}

	totalExpenses, totalIncome := CalculateTotals(convertToIncomeExpense(incExpList))

	report := fmt.Sprintf("Vehicle %d Income/Expense Report\n", vehicleID)
	report += fmt.Sprintf("Start Date: %s\n", startDate.Format("2006-01-02"))
	report += fmt.Sprintf("End Date: %s\n", endDate.Format("2006-01-02"))
	report += fmt.Sprintf("Total Expenses: $%d\n", totalExpenses)
	report += fmt.Sprintf("Total Income: $%d\n", totalIncome)
	report += fmt.Sprintf("Net Income: $%d\n", totalIncome-totalExpenses)

	return c.SendString(report)
}

func convertToIncomeExpense(trips []models.Incomexpe) []IncomeExpense {
	var res []IncomeExpense
	for _, t := range trips {
		res = append(res, IncomeExpense{
			IeAmount:    t.IeAmount,
			IeDate:      t.IeDate.Format("2006-01-02"),
			IsExpense:   !t.IsIncome, // invert the value of IeIsIncome to get the IsExpense field
			IeVehicleID: t.IeVehicleID,
			IeType:      t.IeType,
		})

	}
	return res
}

func GetIncExpList(startDate time.Time, endDate time.Time, vehicleID int) ([]models.Incomexpe, error) {
	incExpList := []models.Incomexpe{}
	err := database.Database.Db.Where("ie_vehicle_id = ? AND ie_date BETWEEN ? AND ?", vehicleID, startDate, endDate).Find(&incExpList).Error
	if err != nil {
		return nil, err
	}
	return incExpList, nil
}

func CalculateTotals(incExpList []IncomeExpense) (int, int) {
	totalIncome := 0
	totalExpenses := 0
	for _, ie := range incExpList {
		if ie.IeType == "income" {
			totalIncome += ie.IeAmount
		} else {
			totalExpenses += ie.IeAmount
		}
	}
	return totalExpenses, totalIncome
}
