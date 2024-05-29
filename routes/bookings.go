package routes

import (
	"errors"
	"fimba/database"
	"fimba/models"
	"hash/fnv"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Booking struct {
	TripID                uint      `json:"t_id"`
	TripCustomerId        uint      `json:"t_customer_id"`
	TripVehicle           string    `json:"t_vechicle"`
	TripType              string    `json:"t_type"`
	TripDriver            string    `json:"t_driver"`
	TripStartDate         time.Time `gorm:"type:timestamp(6)"`
	TripEndDate           time.Time `gorm:"type:timestamp(6)"`
	TripSquadFrmLoc       string    `json:"t_trip_fromlocation"`
	TripSquadToLoc        string    `json:"t_trip_tolocation"`
	TripSquadFrmLat       string    `json:"t_trip_fromlat"`
	TripSquadFrmLng       string    `json:"t_trip_fromlog"`
	TripSquadToLat        string    `json:"t_trip_tolat"`
	TripSquadToLng        string    `json:"t_trip_tolog"`
	TripSquadTotalDist    string    `json:"t_totaldistance"`
	TripSquadAmount       string    `json:"t_trip_amount"`
	TripSquadStatus       string    `json:"t_trip_status"`
	TripSquadTrackingCode string    `json:"t_trackingcode"`
	TripSquadCreated_by   string    `json:"createdBy"`
	CreatedAt             time.Time `gorm:"type:TIMESTAMP(6)"`
	UpdatedAt             time.Time `gorm:"type:TIMESTAMP(6)"`
}

func IsVehicleAvailable(vehicle string, startTime time.Time, endTime time.Time) bool {
	// Retrieve the bookings for the specified vehicle from the database or file, etc.
	bookings, _ := GetBookingsForVehicle(vehicle)

	// Check if the vehicle is available for the requested time period
	for _, booking := range bookings {
		if booking.TripStartDate.Before(endTime) && booking.TripEndDate.After(startTime) {
			return false
		}
	}

	return true
}

func BookVehicle(c *fiber.Ctx) error {
	// Parse the request body into a Booking object

	booking := Booking{}
	if err := c.BodyParser(&booking); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Check if the requested vehicle is available
	vehicleAvailable := IsVehicleAvailable(booking.TripVehicle, booking.TripStartDate, booking.TripEndDate)
	if !vehicleAvailable {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Vehicle is not available for the requested time period"})
	}

	// Book the vehicle
	err := BookVehicleForCustomer(strconv.Itoa(int(booking.TripCustomerId)), booking.TripVehicle, booking.TripType, booking.TripDriver, booking.TripStartDate, booking.TripEndDate)
	if err != nil {
		log.Printf("Error booking vehicle: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Send a response indicating that the vehicle was booked successfully
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Vehicle booked successfully"})
}

func IsVehicleAvailableForEdit(vehicle string, startTimeExisting time.Time, endTimeExisting time.Time, startTimeNew time.Time, endTimeNew time.Time, bookings []Booking, bookingId uint) bool {
	// Check if the vehicle is available for the requested time period
	for _, booking := range bookings {
		if (booking.TripStartDate.Before(endTimeNew) && booking.TripEndDate.After(startTimeNew) && booking.TripID != bookingId) || (booking.TripStartDate.Before(endTimeExisting) && booking.TripEndDate.After(startTimeExisting) && booking.TripID != bookingId) {
			return false
		}
	}

	return true
}

// Helper function to book a vehicle for a customer
func BookVehicleForCustomer(customer string, vehicle string, vehicleType string, driver string, startTime time.Time, endTime time.Time) error {
	// Create a new booking object
	h := fnv.New32a()
	h.Write([]byte(customer))
	booking := &models.Booking{
		TripCustomerId:  uint(h.Sum32()),
		TripVehicle:     vehicle,
		TripType:        vehicleType,
		TripDriver:      driver,
		TripStartDate:   startTime,
		TripEndDate:     endTime,
		TripSquadStatus: "booked",
	}

	// Save the booking to the database or file, etc.
	err := SaveBooking(booking) // pass the address of the booking variable
	if err != nil {
		return err
	}

	return nil
}

// Get the booking by ID
func GetBooking(c *fiber.Ctx) error {
	// Get the booking ID from the URL
	id := c.Params("id")

	// Parse the ID string into a uint
	bookingID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid booking ID",
		})
	}

	// Get the booking from the database
	booking, _ := GetBookingById(uint(bookingID))
	if booking == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Booking not found",
		})
	}

	return c.JSON(booking)
}

func EditBooking(c *fiber.Ctx) error {
	// Parse the request body into a Booking object
	booking := Booking{}
	if err := c.BodyParser(&booking); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Check if the requested booking exists
	existingBooking, _ := GetBookingById(booking.TripID)
	if existingBooking == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Booking not found"})
	}

	// Check if the requested vehicle is available
	vehicleAvailable := IsVehicleAvailableForEdit(existingBooking.TripVehicle, existingBooking.TripStartDate, existingBooking.TripEndDate, booking.TripStartDate, booking.TripEndDate, []Booking{}, existingBooking.TripID)
	if !vehicleAvailable {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Vehicle is not available for the requested time period"})
	}

	// Update the booking
	err := UpdateBooking(int(existingBooking.TripID), booking)
	if err != nil {
		log.Printf("Error updating booking: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Send a response indicating that the booking was updated successfully
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Booking updated successfully"})
}

// Helper function to update a booking
func UpdateBooking(id int, updatedBooking Booking) error {

	// Retrieve the existing booking from the database or file, etc.
	existingBooking, err := GetBookingById(uint(id))
	if err != nil {
		return errors.New("booking not found")
	}
	if existingBooking == nil {
		return errors.New("booking not found")
	}

	// Update the fields of the existing booking
	existingBooking.TripCustomerId = updatedBooking.TripCustomerId
	existingBooking.TripVehicle = updatedBooking.TripVehicle
	existingBooking.TripType = updatedBooking.TripType
	existingBooking.TripDriver = updatedBooking.TripDriver
	existingBooking.TripStartDate = updatedBooking.TripStartDate
	existingBooking.TripEndDate = updatedBooking.TripEndDate
	existingBooking.TripSquadStatus = updatedBooking.TripSquadStatus

	// Save the updated booking to the database or file, etc.

	err = SaveBooking(existingBooking)
	if err != nil {
		return err
	}

	return nil
}

func ListBookings(c *fiber.Ctx) error {
	// Retrieve all bookings from the database or file, etc.
	bookings := GetAllBookings()

	// Return a JSON response with the list of bookings
	return c.JSON(bookings)
}

// Helper function to retrieve all bookings
func GetAllBookings() []Booking {
	// Retrieve all bookings from the database or file, etc.
	// and return them as a slice of Booking objects
	return []Booking{}
}
func GetBookingsForVehicle(vehicleId string) ([]Booking, error) {
	// Get all bookings for the specified vehicle ID
	var bookings []Booking
	if err := database.Database.Db.Where("vehicle_id = ?", vehicleId).Find(&bookings).Error; err != nil {
		return nil, err
	}

	// Return the bookings
	return bookings, nil
}

// Define the SaveBooking function
func SaveBooking(booking *models.Booking) error {
	result := database.Database.Db.Create(&booking)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func GetBookingById(id uint) (*models.Booking, error) {
	var booking models.Booking

	// Retrieve the booking from the database or file, etc.
	err := database.Database.Db.Find(&booking, id).Error
	if err != nil {
		return nil, err
	}

	return &booking, nil
}
