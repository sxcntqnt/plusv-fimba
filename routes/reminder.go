package routes

import (
	"fimba/database"
	"fimba/models"

	"github.com/gofiber/fiber/v2"

	"log"
	"time"
)

type Reminder struct {
	RmdId     uint      `json:"r_id"`
	RmdDate   time.Time `gorm:"timestamp(6)"`
	RmdMsg    string    `json:"r_message"`
	RmdIsRead string    `json:"r_isread"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6)"`
	VehicleID uint      // Foreign key for Vehicle model
	Vehicle   Vehicle   // Associated Vehicle model
}

func CreateReminder(c *fiber.Ctx) error {
	// Parse the request body into a Reminder object
	reminder := models.Reminder{}
	if err := c.BodyParser(&reminder); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Save the reminder to the database or file, etc.
	err := SaveReminder(reminder)
	if err != nil {
		log.Printf("Error saving reminder: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Send a response indicating that the reminder was created successfully
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Reminder created successfully"})
}

// Helper function to save a reminder to the database or file, etc.
func SaveReminder(reminder models.Reminder) error {
	// Save the reminder to the database using GORM's Save method
	err := database.Database.Db.Save(&reminder).Error
	if err != nil {
		return err
	}

	// Return nil if the operation was successful
	return nil
}

// Helper function to update a reminder in a database or file, etc.
func UpdateReminder(reminder *models.Reminder) error { // Fix here - change parameter type to *models.Reminder
	// Update the reminder in the database using GORM's Save method
	err := database.Database.Db.Save(reminder).Error // Fix here - remove the & from reminder
	if err != nil {
		return err
	}

	// Return nil error if the operation was successful
	return nil
}

func EditReminder(c *fiber.Ctx) error {
	// Parse the request body into a Reminder object
	reminder := models.Reminder{}
	if err := c.BodyParser(&reminder); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Get the reminder ID from the request parameters
	reminderID := c.Params("id")

	// Retrieve the reminder with the given ID from the database or file, etc.
	storedReminder, err := RetrieveReminder(reminderID)
	if err != nil {
		log.Printf("Error retrieving reminder: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Update the stored reminder with the new values
	storedReminder.Vehicle = reminder.Vehicle
	storedReminder.RmdDate = reminder.RmdDate
	storedReminder.RmdMsg = reminder.RmdMsg

	// Store the updated reminder in the database or file, etc.
	err = UpdateReminder(storedReminder) // Fix here - pass storedReminder directly without &
	if err != nil {
		log.Printf("Error updating reminder: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Send a response indicating that the reminder was updated successfully
	return c.Status(fiber.StatusOK).JSON(storedReminder)
}

func RetrieveReminder(reminderID string) (*models.Reminder, error) { // Fix here - change return type to *models.Reminder
	// Create an empty Reminder struct to hold the retrieved reminder
	reminder := &models.Reminder{}

	// Retrieve the reminder with the given ID from the database using GORM's Find method
	err := database.Database.Db.Where("id = ?", reminderID).First(reminder).Error
	if err != nil {
		return nil, err
	}

	// Return the retrieved reminder and nil error if the operation was successful
	return reminder, nil
}

func GetReminder(c *fiber.Ctx) error {
	// Get the reminder ID from the request parameters
	reminderID := c.Params("id")

	// Retrieve the reminder with the given ID from the database or file, etc.
	storedReminder, err := RetrieveReminder(reminderID)
	if err != nil {
		log.Printf("Error retrieving reminder: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Create a new instance of the Reminder struct
	reminderResult := models.Reminder{}

	// Copy the retrieved reminder data into the reminderResult
	reminderResult.RmdId = storedReminder.RmdId
	reminderResult.RmdMsg = storedReminder.RmdMsg
	reminderResult.RmdDate = storedReminder.RmdDate
	reminderResult.RmdIsRead = storedReminder.RmdIsRead
	reminderResult.CreatedAt = storedReminder.CreatedAt

	// Retrieve the associated vehicle for the reminder using GORM's Preload method
	err = database.Database.Db.Model(storedReminder).Preload("Vehicle").Error
	if err != nil {
		return err
	}

	// Set the associated vehicle to the reminderResult
	reminderResult.Vehicle = storedReminder.Vehicle

	// Return the retrieved reminder as JSON
	return c.Status(fiber.StatusOK).JSON(reminderResult)
}
func ListReminders(c *fiber.Ctx) error {
	reminders := []models.Reminder{}
	if err := database.Database.Db.Find(&reminders).Error; err != nil {
		log.Printf("Error querying database: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Create a slice to store the reminders
	remindersSlice := make([]Reminder, 0)

	// Iterate over the query results and append each reminder to the slice
	for _, reminder := range reminders {
		// Convert models.Reminder to Reminder, including only the Vehicle ID
		reminderResult := Reminder{
			RmdId:     reminder.RmdId,
			RmdDate:   reminder.RmdDate,
			RmdMsg:    reminder.RmdMsg,
			RmdIsRead: reminder.RmdIsRead,
			CreatedAt: reminder.CreatedAt,
			Vehicle: Vehicle{
				V_Id: reminder.Vehicle.V_Id,
			},
		}
		remindersSlice = append(remindersSlice, reminderResult)
	}

	// Send the reminders slice as a JSON response
	return c.Status(fiber.StatusOK).JSON(remindersSlice)
}
