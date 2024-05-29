package routes

import (
	"fimba/database"
	"fimba/models"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

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

func CreateFuelEntry(c *fiber.Ctx) error {
	// Parse the request body into a FuelEntry object
	var fuelEntry models.FuelEntry

	if err := c.BodyParser(&fuelEntry); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Save the fuel entry to the database or file, etc.
	vehicleID := uint(fuelEntry.VehicleID)
	savedFuelEntry, err := SaveFuelEntry(&fuelEntry, vehicleID)
	if err != nil {
		log.Printf("Error saving fuel entry: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Send a response indicating that the fuel entry was created successfully
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Fuel entry created successfully", "fuelEntry": savedFuelEntry})
}

// Helper function to save a fuel entry to the database
// SaveFuelEntry saves a fuel entry to the database
func SaveFuelEntry(fuelEntry *models.FuelEntry, vehicleID uint) (*models.FuelEntry, error) {
	fuelEntry.VehicleID = int(vehicleID) // Update VehicleID field with the provided vehicleID
	result := database.Database.Db.Save(fuelEntry)
	if result.Error != nil {
		return nil, result.Error
	}
	return fuelEntry, nil
}

func EditFuelEntry(c *fiber.Ctx) error {
	// Parse the request body into a FuelEntry object
	fuelEntry := FuelEntry{}
	if err := c.BodyParser(&fuelEntry); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Update the fuel entry in the database or file, etc.
	err := UpdateFuelEntry(&models.FuelEntry{ // Pass a pointer to models.FuelEntry with updated fields
		Id: fuelEntry.Id,
		// Update other fields as needed
	})
	if err != nil {
		log.Printf("Error updating fuel entry: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Send a response indicating that the fuel entry was updated successfully
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Fuel entry updated successfully"})
}

// Helper function to update a fuel entry in the database or file, etc.
func UpdateFuelEntry(fuelEntry *models.FuelEntry) error { // Update the argument type here
	// Update the fuel entry in the database
	result := database.Database.Db.Model(fuelEntry).Where("id = ?", fuelEntry.Id).Updates(fuelEntry) // Update the argument here
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func ListFuelEntries(c *fiber.Ctx) error {
	// Query the fuel entries from the database
	var fuelEntries []FuelEntry
	result := database.Database.Db.Find(&fuelEntries)
	if result.Error != nil {
		log.Printf("Error querying database: %v\n", result.Error)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	// Send a response containing the list of fuel entries
	return c.Status(fiber.StatusOK).JSON(fuelEntries)
}

// addToExpense adds the fuel entry to the expenses
func AddToExpense(c *fiber.Ctx, fuelEntry *models.FuelEntry) error { // Update the argument type here
	if err := c.BodyParser(&fuelEntry); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Save or update the fuel entry in the database
	if fuelEntry.Id == 0 {
		// If ID is 0, it's a new fuel entry, so save it
		savedFuelEntry, err := SaveFuelEntry(fuelEntry, uint(fuelEntry.VehicleID))
		if err != nil {
			log.Printf("Error saving fuel entry: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Add the fuel entry to expenses
		err = AddToExpense(c, savedFuelEntry)
		if err != nil {
			log.Printf("Error adding fuel entry to expenses: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Send a response indicating that the fuel entry was added to expenses successfully
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Fuel entry added to expenses successfully", "FuelEntry": savedFuelEntry})
	} else {
		// If ID is not 0, it's an existing fuel entry, so update it
		err := UpdateFuelEntry(fuelEntry) // Fix the argument type here
		if err != nil {
			log.Printf("Error updating fuel entry: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Add the fuel entry to expenses
		err = AddToExpense(c, fuelEntry)
		if err != nil {
			log.Printf("Error adding fuel entry to expenses: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Send a response indicating that the fuel entry was updated successfully
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Fuel entry updated successfully", "FuelEntry": fuelEntry})
	}
}
