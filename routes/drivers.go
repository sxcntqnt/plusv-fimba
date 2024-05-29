package routes

import (
	"errors"
	"fimba/database"
	"fimba/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Driver struct {
	//this is the serializer
	DID              uint      `json:"d_id"`
	DName            string    `json:"d_name"`
	DMobile          string    `json:"d_mobile"`
	DAddress         string    `json:"d_address"`
	DAge             int       `json:"d_age"`
	DLicenseNo       string    `json:"d_licenseno"`
	DLicenceExpdate  time.Time `gorm:"type:TIMESTAMP(6)"`
	DTotalExperience int       `json:"d_total_exp"`
	DDateOfJoining   time.Time `gorm:"type:TIMESTAMP(6)"`
	DReference       string    `json:"d_ref"`
	DIsActive        int       `json:"d_is_active"`
	DCreatedBy       uint      `json:"d_created_by"`
}

func CreateResponseDriver(driverModel models.Driver) Driver {
	return Driver{DID: driverModel.DID, DName: driverModel.DName, DMobile: driverModel.DMobile, DAddress: driverModel.DAddress, DAge: driverModel.DAge, DLicenseNo: driverModel.DLicenseNo, DLicenceExpdate: driverModel.DLicenceExpdate, DTotalExperience: driverModel.DTotalExperience, DDateOfJoining: driverModel.DDateOfJoining, DReference: driverModel.DReference, DIsActive: driverModel.DIsActive, DCreatedBy: driverModel.DCreatedBy}
}

func CreateDriver(c *fiber.Ctx) error {
	var driver models.Driver

	if err := c.BodyParser(&driver); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	database.Database.Db.Create(&driver)
	responseDriver := CreateResponseDriver(driver)

	return c.Status(200).JSON(responseDriver)
}
func GetDrivers(c *fiber.Ctx) error {
	drivers := []models.Driver{}

	database.Database.Db.Find(&drivers)
	responseDrivers := []Driver{}
	for _, driver := range drivers {
		responseDriver := CreateResponseDriver(driver)
		responseDrivers = append(responseDrivers, responseDriver)
	}
	return c.Status(200).JSON(responseDrivers)
}
func findDriver(id int, Driver *models.Driver) error {
	database.Database.Db.Find(&Driver, "id = ?", id)
	if Driver.DID == 0 {
		return errors.New("Driver does not exist")
	}
	return nil
}
func GetDriver(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var driver models.Driver

	if err != nil {
		return c.Status(400).JSON("Please ensure that :id is an integer")
	}
	if err := findDriver(id, &driver); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	responseDriver := CreateResponseDriver(driver)
	return c.Status(200).JSON(responseDriver)

}

func UpdateDriver(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var driver models.Driver

	if err != nil {
		return c.Status(400).JSON("Please ensure that :id is an integer")
	}
	if err := findDriver(id, &driver); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	type UpdateDriver struct {
		DName            string    `json:"d_name"`
		DMobile          string    `json:"d_mobile"`
		DAddress         string    `json:"d_address"`
		DLicenceExpdate  time.Time `gorm:"type:TIMESTAMP(6)"`
		DTotalExperience int       `json:"d_total_exp"`
		DDateOfJoining   time.Time `gorm:"type:TIMESTAMP(6)"`
		DReference       string    `json:"d_ref"`
		DIsActive        int       `json:"d_is_active"`
	}

	var updateData UpdateDriver
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(500).JSON(err.Error())
	}

	driver.DName = updateData.DName
	driver.DMobile = updateData.DMobile
	driver.DAddress = updateData.DAddress
	driver.DLicenceExpdate = updateData.DLicenceExpdate
	driver.DDateOfJoining = updateData.DDateOfJoining
	driver.DReference = updateData.DReference
	driver.DIsActive = updateData.DIsActive

	database.Database.Db.Save(&driver)
	responseDriver := CreateResponseDriver(driver)
	return c.Status(200).JSON(responseDriver)
}
func DeleteDriver(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var driver models.Driver

	if err != nil {
		return c.Status(400).JSON("Please ensure the :id is an integer")
	}
	if err := findDriver(id, &driver); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	if err := database.Database.Db.Delete(&driver).Error; err != nil {
		return c.Status(404).JSON(err.Error())
	}
	return c.Status(200).SendString("Sucessfully Deleted Driver ")
}
