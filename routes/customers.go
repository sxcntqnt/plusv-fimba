package routes

import (
	"errors"
	"fimba/database"
	"fimba/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Customer struct {
	Id            uint      `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email" "gorm:unique"`
	Address       string    `json:"address"`
	CreatedAt     time.Time `gorm:"type:TIMESTAMP(6)"`
	IsActive      string    `json:"isactive"`
	Modified_date time.Time `gorm:"type:TIMESTAMP(6)"`
	Password      []byte    `"json:"-"`
}

func CreateResponseCustomer(customerModel models.Customer) Customer {
	return Customer{Id: customerModel.Id, Name: customerModel.Name, Email: customerModel.Email, Address: customerModel.Address, IsActive: customerModel.IsActive, Modified_date: customerModel.Modified_date, Password: customerModel.Password}
}

func CreateCustomer(c *fiber.Ctx) error {
	var customer models.Customer

	if err := c.BodyParser(&customer); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	database.Database.Db.Create(&customer)
	responseCustomer := CreateResponseCustomer(customer)

	return c.Status(200).JSON(responseCustomer)
}
func GetCustomers(c *fiber.Ctx) error {
	Customers := []models.Customer{}

	database.Database.Db.Find(&Customers)
	responseCustomers := []Customer{}
	for _, Customer := range Customers {
		responseCustomer := CreateResponseCustomer(Customer)
		responseCustomers = append(responseCustomers, responseCustomer)
	}
	return c.Status(200).JSON(responseCustomers)
}

func findCustomer(id int, customer *models.Customer) error {
	database.Database.Db.Find(&customer, "id = ?", id)
	if customer.Id == 0 {
		return errors.New("Customer does not exist")
	}
	return nil
}
func GetCustomer(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var customer models.Customer

	if err != nil {
		return c.Status(400).JSON("Please ensure that :id is an integer")
	}
	if err := findCustomer(id, &customer); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	responseCustomer := CreateResponseCustomer(customer)
	return c.Status(200).JSON(responseCustomer)
}

func UpdateCustomer(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var customer models.Customer

	if err != nil {
		return c.Status(400).JSON("Please ensure that :id is an integer")
	}
	if err := findCustomer(id, &customer); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	type UpdateCustomer struct {
		Name     string `json:"name"`
		Email    string `json:"email" "gorm:unique"`
		Address  string `json:"address"`
		IsActive string `json:"isactive"`
		Password []byte `"json:"-"`
	}

	var updateData UpdateCustomer
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(500).JSON(err.Error())
	}

	customer.Name = updateData.Name
	customer.Email = updateData.Email
	customer.Address = updateData.Address
	customer.IsActive = updateData.IsActive
	customer.Password = updateData.Password

	database.Database.Db.Save(&customer)

	responseCustomer := CreateResponseCustomer(customer)
	return c.Status(200).JSON(responseCustomer)
}

func DeleteCustomer(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var customer models.Customer

	if err != nil {
		return c.Status(400).JSON("Please ensure the :id is an integer")
	}
	if err := findCustomer(id, &customer); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	if err := database.Database.Db.Delete(&customer).Error; err != nil {
		return c.Status(404).JSON(err.Error())
	}
	return c.Status(200).SendString("Sucessfully Deleted Customer ")
}
