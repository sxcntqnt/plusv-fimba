package routes

import (
	"errors"
	"fimba/database"
	"fimba/models"

	"github.com/gofiber/fiber/v2"

	"time"
)

type Incomexpe struct {
	IeID          uint      `json:"ie_id"`
	IeVehicleID   string    `json:"ie_v_id"`
	IeDate        time.Time `json:"ie_date"`
	IeType        string    `json:"ie_type"`
	IeDescription string    `json:"ie_description"`
	IeAmount      int       `json:"ie_amount"`
	CreatedAt     time.Time `gorm:"type:TIMESTAMP(6)"`
	UpdatedAt     time.Time `gorm:"type:TIMESTAMP(6)"`
}

func CreateResponseIncome(IncModel models.Incomexpe) Incomexpe {
	return Incomexpe{IeID: IncModel.IeID, IeVehicleID: IncModel.IeVehicleID, IeDate: IncModel.IeDate, IeType: IncModel.IeType, IeDescription: IncModel.IeDescription, IeAmount: IncModel.IeAmount}
}

func AddIncExp(c *fiber.Ctx) error {
	var Incomexpe models.Incomexpe

	if err := c.BodyParser(&Incomexpe); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	// Validate the options
	if Incomexpe.IeID == 0 || Incomexpe.IeVehicleID == "" || !Incomexpe.IeDate.IsZero() || (Incomexpe.IeType != "income" && Incomexpe.IeType == "expense") || Incomexpe.IeDescription == "" || Incomexpe.IeAmount == -1 {
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	database.Database.Db.Create(&Incomexpe)
	responseIncome := CreateResponseIncome(Incomexpe)

	return c.Status(200).JSON(responseIncome)
}
func GetIncomes(c *fiber.Ctx) error {
	Incomes := []models.Incomexpe{}

	database.Database.Db.Find(&Incomes)
	responseIncomes := []Incomexpe{}
	for _, Incomexpe := range Incomes {
		responseIncome := CreateResponseIncome(Incomexpe)
		responseIncomes = append(responseIncomes, responseIncome)
	}
	return c.Status(200).JSON(responseIncomes)
}
func FindIncomes(id int, Incomexpe *models.Incomexpe) error {
	database.Database.Db.Find(&Incomexpe, "id = ?", id)
	if Incomexpe.IeID == 0 {
		return errors.New("Income doesnt exist")
	}
	return nil

}
func GetIncome(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var Incomexpe models.Incomexpe

	if err != nil {
		return c.Status(400).JSON("Please ensure that :id is an integer")
	}
	if err := FindIncomes(id, &Incomexpe); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	responseIncome := CreateResponseIncome(Incomexpe)
	return c.Status(200).JSON(responseIncome)
}

func EditIncome(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var Incomexpe models.Incomexpe

	if err != nil {
		return c.Status(400).JSON("Please ensure that :id is an integer")
	}
	if err := FindIncomes(id, &Incomexpe); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	type UpdateIncome struct {
		IeVehicleID   string    `json:"ie_v_id"`
		IeDate        time.Time `json:"ie_date"`
		IeType        string    `json:"ie_type"`
		IeDescription string    `json:"ie_description"`
		IeAmount      int       `json:"ie_amount"`
	}

	var updateRecord UpdateIncome
	if err := c.BodyParser(&updateRecord); err != nil {
		return c.Status(500).JSON(err.Error())
	}

	Incomexpe.IeVehicleID = updateRecord.IeVehicleID
	Incomexpe.IeDate = updateRecord.IeDate
	Incomexpe.IeType = updateRecord.IeType
	Incomexpe.IeDescription = updateRecord.IeDescription
	Incomexpe.IeAmount = updateRecord.IeAmount

	database.Database.Db.Save(&Incomexpe)
	responseIncome := CreateResponseIncome(Incomexpe)
	return c.Status(200).JSON(responseIncome)
}
