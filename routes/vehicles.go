package routes

import (
	"errors"
	"fimba/database"
	"fimba/models"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Vehicle struct {
	V_Id              uint        `json:"v_id" gorm:"primaryKey"`
	V_RegistrationNo  string      `json:"v__registration_no"`
	V_Name            string      `json:"v_nameo"`
	V_Model           string      `json:"v_modelo"`
	V_ChassisNo       string      `json:"v_chassis_noo"`
	V_EngineNo        string      `json:"v_engine_noo"`
	V_Manufactured_by string      `json:"v_manufactured_byo"`
	V_Type            string      `json:"v_typeo"`
	V_Color           string      `json:"v_coloro"`
	V_Mileageperlitre string      `json:"v_mileageperlitreo"`
	V_Is_active       int         `json:"v_is_activeo"`
	V_Group           int         `json:" v_groupo"`
	V_Reg_exp_date    string      `json:"v_reg_exp_dateo"`
	V_Api_url         string      `json:"v_api_urlo"`
	V_Api_username    string      `json:"v_api_usernameo"`
	V_Api_password    string      `json:"v_api_passwordo"`
	V_Created_by      string      `json:"v_created_byo"`
	V__CreatedAt      time.Time   `gorm:"type:timestamp(6)"`
	V_modified_date   time.Time   `gorm:"type:timestamp(6)"`
	FuelEntries       []FuelEntry `gorm:"foreignKey:VehicleID"`
}

type VehicleGroups struct {
	GroupID   uint      `json:"gr_id"`
	GroupName string    `json:"gr_name"`
	GroupDesc string    `json:"gr_desc"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP(6)"`

	// Define a one-to-many relationship between VehicleGroup and Vehicle
	Vehicles []Vehicle `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func CreateResponseVehicle(VehicleModel models.Vehicle) Vehicle {
	return Vehicle{V_Id: VehicleModel.V_Id, V_RegistrationNo: VehicleModel.V_RegistrationNo, V_Name: VehicleModel.V_Name, V_Model: VehicleModel.V_Model, V_ChassisNo: VehicleModel.V_ChassisNo, V_EngineNo: VehicleModel.V_EngineNo, V_Manufactured_by: VehicleModel.V_Manufactured_by, V_Type: VehicleModel.V_Type, V_Color: VehicleModel.V_Color, V_Mileageperlitre: VehicleModel.V_Mileageperlitre, V_Is_active: VehicleModel.V_Is_active, V_Group: VehicleModel.V_Group, V_Reg_exp_date: VehicleModel.V_Reg_exp_date, V_Api_url: VehicleModel.V_Api_url, V_Api_username: VehicleModel.V_Api_username, V_Api_password: VehicleModel.V_Api_password, V_Created_by: VehicleModel.V_Created_by}
}

func CreateVehicle(c *fiber.Ctx) error {
	var Vehicle models.Vehicle

	if err := c.BodyParser(&Vehicle); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	database.Database.Db.Create(&Vehicle)
	responseVehicle := CreateResponseVehicle(Vehicle)

	return c.Status(200).JSON(responseVehicle)
}
func GetVehicles(c *fiber.Ctx) error {
	Vehicles := []models.Vehicle{}

	database.Database.Db.Find(&Vehicles)
	responseVehicles := []Vehicle{}
	for _, Vehicle := range Vehicles {
		responseVehicle := CreateResponseVehicle(Vehicle)
		responseVehicles = append(responseVehicles, responseVehicle)
	}
	return c.Status(200).JSON(responseVehicles)
}
func findVehicle(id int, Vehicle *models.Vehicle) error {
	database.Database.Db.Find(&Vehicle, "id = ?", id)
	if Vehicle.V_Id == 0 {
		return errors.New("Vehicle does not exist")
	}
	return nil
}
func GetVehicle(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var Vehicle models.Vehicle

	if err != nil {
		return c.Status(200).JSON("Please ensure that :id is an integer")
	}
	if err := findVehicle(id, &Vehicle); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	responseVehicle := CreateResponseVehicle(Vehicle)
	return c.Status(200).JSON(responseVehicle)
}

func UpdateVehicle(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var Vehicle models.Vehicle

	if err != nil {
		return c.Status(400).JSON("Please ensure that :id is an integer")
	}
	if err := findVehicle(id, &Vehicle); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	type UpdateVehicle struct {
	}

	var updateVehicle UpdateVehicle
	if err := c.BodyParser(&updateVehicle); err != nil {
		return c.Status(500).JSON(err.Error())
	}

	database.Database.Db.Save(&Vehicle)
	responseVehicle := CreateResponseVehicle(Vehicle)
	return c.Status(200).JSON(responseVehicle)
}
func DeleteVehicle(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var Vehicle models.Vehicle

	if err != nil {
		return c.Status(400).JSON("Please ensure the :id is an integer")
	}
	if err := findVehicle(id, &Vehicle); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	if err := database.Database.Db.Delete(&Vehicle).Error; err != nil {
		return c.Status(404).JSON(err.Error())
	}
	return c.Status(200).SendString("Sucessfully Deleted Vehicle ")
}

func CreateVehicleGroup(c *fiber.Ctx) error {
	// Parse the request body into a new VehicleGroup object
	group := new(VehicleGroups)
	if err := c.BodyParser(group); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Do some validation on the input
	if group.GroupName == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Group name cannot be empty")
	}
	if len(group.Vehicles) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("At least one vehicle is required in the group")
	}

	// Store the group in a database or file, etc.
	log.Printf("Created new vehicle group '%s' with %d vehicles\n", group.GroupName, len(group.Vehicles))

	// Send a response indicating that the group was created
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Group created successfully"})
}

func AddVehiclesToGroup(c *fiber.Ctx) error {
	// Get the name of the group to add vehicles to from the URL parameter
	groupName := c.Params("groupName")

	// Parse the request body into a slice of Vehicle objects
	vehiclesToAdd := []models.Vehicle{}
	if err := c.BodyParser(&vehiclesToAdd); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Retrieve the existing group from the database
	var existingGroup models.VehicleGroups

	if err := database.Database.Db.Where("name = ?", groupName).First(&existingGroup).Error; err != nil {
		log.Printf("Error retrieving group from database: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Add the new vehicles to the group
	for _, vehicle := range vehiclesToAdd {
		existingGroup.Vehicles = append(existingGroup.Vehicles, vehicle)
	}

	// Save the updated group back to the database
	if err := database.Database.Db.Save(&existingGroup).Error; err != nil {
		log.Printf("Error saving group to database: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Send a response indicating that the vehicles were added to the group
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Vehicles added to group successfully"})
}

// Helper function to delete specified vehicles from a slice of vehicles
func DeleteVehicles(vehicles *[]models.Vehicle, vehiclesToDelete []models.Vehicle) []models.Vehicle {
	// Create a map of the vehicles to delete for efficient lookup
	vehiclesToDeleteMap := make(map[uint]bool)
	for _, v := range vehiclesToDelete {
		vehiclesToDeleteMap[v.V_Id] = true
	}

	// Create a new slice of vehicles without the vehicles to delete
	updatedVehicles := []models.Vehicle{}
	for _, v := range *vehicles {
		if _, ok := vehiclesToDeleteMap[v.V_Id]; !ok {
			updatedVehicles = append(updatedVehicles, v)
		}
	}

	return updatedVehicles
}

func SaveVehicleGroup(group *models.VehicleGroups) error {
	// Save the group to the database
	result := database.Database.Db.Save(group)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetVehicleGroup is a function that retrieves a vehicle group by ID
func GetVehicleGroup(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON("Please ensure the :id is an integer")
	}

	var group models.VehicleGroups
	if err := database.Database.Db.First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no group found, return appropriate response
			return errors.New("Vehicle Group not found")
		}
		return c.Status(500).JSON(err.Error())
	}

	return c.JSON(group)
}

func DeleteVehiclesFromGroup(c *fiber.Ctx) error {
	// Parse the request body into a slice of models.Vehicle objects to delete
	vehiclesToDelete := []models.Vehicle{}
	if err := c.BodyParser(&vehiclesToDelete); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Retrieve the existing group from the database or file, etc.
	existingGroup := new(models.VehicleGroups)
	err := GetVehicleGroupByName(c)
	if err != nil {
		log.Printf("Error retrieving group from database: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Delete the specified vehicles from the group
	existingGroup.Vehicles = DeleteVehicles(&existingGroup.Vehicles, vehiclesToDelete)

	// Store the updated group back in the database or file, etc.
	if err := SaveVehicleGroup(existingGroup); err != nil {
		log.Printf("Error saving group to database: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Send a response indicating that the vehicles were deleted from the group
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Vehicles deleted from group"})
}

// GetVehicleGroupByName is a function that retrieves a vehicle group by name
func GetVehicleGroupByName(c *fiber.Ctx) error {
	name := c.Params("name") // Get the name parameter from the URL

	// Query the database to find the VehicleGroup by name
	group := &models.VehicleGroups{}
	if err := database.Database.Db.Where("gr_name = ?", name).First(group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no group found, return appropriate response
			return errors.New("Vehicle Group not found")
		}
		// If there's any other error, return error response
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// If group found, return it as JSON response
	return c.JSON(group)
}

// GetVehicleGroupByID is a function that retrieves a vehicle group by ID from the SQLite database
func GetVehicleGroupByID(id int, group *models.VehicleGroups) error {
	if err := database.Database.Db.Where("gr_id = ?", id).First(group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no group found, return appropriate response
			return errors.New("Vehicle Group not found")
		}
		return err
	}
	return nil
}

// routes.UpdateVehicleGroup is a function that updates a vehicle group by ID
func UpdateVehicleGroup(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON("Please ensure the :id is an integer")
	}

	var group models.VehicleGroups
	if err := GetVehicleGroupByID(id, &group); err != nil {
		return c.Status(404).JSON(err.Error())
	}

	// Parse request body into a new vehicle group object
	newGroup := new(models.VehicleGroups)
	if err := c.BodyParser(newGroup); err != nil {
		return c.Status(400).JSON("Failed to parse request body")
	}

	// Update the fields of the existing group with the new group object
	group.GroupName = newGroup.GroupName
	group.GroupDesc = newGroup.GroupDesc

	// Save the updated group to the database
	if err := database.Database.Db.Save(&group).Error; err != nil {
		return c.Status(500).JSON(err.Error())
	}

	return c.Status(200).JSON(group)
}
