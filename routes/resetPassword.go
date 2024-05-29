package routes

import (
	"fimba/database"
	"fimba/models"

	"github.com/gofiber/fiber/v2"
)

type ResetPasswordRequest struct {
	Password string `json:"password"`
}

func ResetPassword(c *fiber.Ctx) error {
	// Parse the request body into a ResetPasswordRequest object
	newPassword := ResetPasswordRequest{}
	if err := c.BodyParser(&newPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	// Get the user ID from the request parameters
	userID := c.Params("id")

	// Update the user's password in the database
	if err := database.Database.Db.Model(&models.LoginData{}).Where("id = ?", userID).Update("password", newPassword.Password).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	// Send a response indicating that the password was updated successfully
	return c.Status(fiber.StatusOK).SendString("Password updated successfully")
}
