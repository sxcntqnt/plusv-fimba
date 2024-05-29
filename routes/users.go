package routes

import (
	"errors"
	"fimba/database"
	"fimba/models"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const SecretKey = "d779bb9cb37697c541987d1c9c46157afcc9712d1288043a8c67da89f186cafe"

type User struct {
	UId         uint   `json:"uid"`
	UName       string `json:"uname"`
	UUsername   string `json:"uusername"`
	UEmail      string `json:"email" gorm:"unique"`
	UPassword   string `json:"-"` // the dash indicates that this field should not be JSON-encoded
	URole       string `json:"role"`
	UIsActive   bool   `json:"active"`
	Permissions *models.Login_roles
	LoginData       *models.LoginData // new field for login information
	UCreatedAt  time.Time
}
type Login struct {
	Number   string `json:"number" gorm "unique"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
	Role     string `json:"role"`
	APIKey   string `json:"api_key"`
}

func CreateResponseUser(userModel models.User) User {
	return User{UId: userModel.UId, UName: userModel.UName, UUsername: userModel.UUsername, UPassword: userModel.UPassword, UEmail: userModel.UEmail, UIsActive: userModel.UIsActive}
}
func RegisterUser(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Error parsing user data"})
	}

	// Validate input data
	if user.UName == "" || user.UUsername == "" || user.UEmail == "" || len(user.UPassword) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Missing required fields"})
	}
	if user.UPassword == "" || len(user.UPassword) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Password is required"})
	}

	if !isEmailValid(user.UEmail) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid email address"})
	}

	// check if user with the given email already exists
	var count int64
	if err := database.Database.Db.Find(&User{}).Where("email = ?", user.UEmail).Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error checking user existence"})
	}
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "User with the given email already exists"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.UPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error hashing password"})
	}

	// create new user
	newUser := &User{
		UName:     user.UName,
		UUsername: user.UUsername,
		UEmail:    user.UEmail,
		UPassword: string(hashedPassword),
		URole:     user.URole,
		UIsActive: true,
		Permissions: &models.Login_roles{
			LrVehicleList:        1,
			LrVehicleListView:    1,
			LrVehicleListEdit:    1,
			LrVehicleAdd:         1,
			LrVehicleGroup:       1,
			LrVehicleGroupAdd:    1,
			LrVehicleGroupAction: 1,
			LrDriversList:        1,
			LrDriversListEdit:    1,
			LrDriversAdd:         1,
			LrTripsList:          1,
			LrTripsListEdit:      1,
			LrTripsAdd:           1,
			LrCustomerList:       1,
			LrCustomerEdit:       1,
			LrCustomerAdd:        1,
			LrFuelList:           1,
			LrFuelEdit:           1,
			LrFuelAdd:            1,
			LrReminderList:       1,
			LrReminderDelete:     1,
			LrReminderAdd:        1,
			LrIeList:             1,
			LrIeEdit:             1,
			LrIeAdd:              1,
			LrIeTracking:         1,
			LrLiveLoc:            1,
			LrGeofenceAdd:        1,
			LrGeofenceList:       1,
			LrGeofenceDelete:     1,
			LrGeofenceEvents:     1,
			LrReports:            1,
			LrSettings:           1,
		},
		LoginData: &models.LoginData{
			Email:    user.UEmail,
			Password: user.UPassword,
			Role:     user.URole,
			APIKey:   user.LoginData.APIKey,
		},
	}

	// Save user to database
	if err := database.Database.Db.Create(newUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error saving user to database"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User created successfully"})
}

// Utility function to check if an email address is valid
func isEmailValid(email string) bool {
	// This is a simple regex pattern to check if the email is valid
	// A more comprehensive check can be done depending on the requirements
	pattern := "^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$"
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

func findUser(id int, user *models.User) error {
	database.Database.Db.Find(&user, "id = ?", id)
	if user.UId == 0 {
		return errors.New("Customer does not exist")
	}
	return nil
}

func UpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var user models.User

	if err != nil {
		return c.Status(400).JSON("Please ensure that :id is an integer")
	}
	if err := findUser(id, &user); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	type UpdateUser struct {
		UName     string `json:"u_name"`
		UUsername string `json:"u_username"`
		UPassword []byte `json:"u_password"`
		UIsActive bool   `json:"u_is_active"`
		UEmail    string `json:"u_email"`
	}

	var updateData UpdateUser
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(500).JSON(err.Error())
	}

	if len(updateData.UPassword) > 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword(updateData.UPassword, 14)
		if err != nil {
			return c.Status(500).JSON(err.Error())
		}
		user.UPassword = string(hashedPassword)
	}

	user.UName = updateData.UName
	user.UUsername = updateData.UUsername
	user.UIsActive = updateData.UIsActive
	user.UEmail = updateData.UEmail

	database.Database.Db.Save(&user)
	responseUser := CreateResponseUser(user)
	return c.Status(200).JSON(responseUser)
}

func DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	var user models.User

	if err != nil {
		return c.Status(400).JSON("Please ensure the :id is an integer")
	}
	if err := findUser(id, &user); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	if err := database.Database.Db.Delete(&user).Error; err != nil {
		return c.Status(404).JSON(err.Error())
	}
	return c.Status(200).SendString("Sucessfully Deleted Customer ")
}

// User role is not defined.Please contact admin'
func CurrentUser(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	claims, err := VerifyToken(cookie)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token",
		})
	}

	var user models.LoginData
	err = database.Database.Db.Where("id = ?", claims.Issuer).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to retrieve user from the database",
			})
		}
	}

	return c.JSON(user)
}
func VerifyToken(tokenString string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return nil, errors.New("Invalid token claims")
	}

	return claims, nil
}
func LoginUser(c *fiber.Ctx, email string) (*models.User, error) {
	var user models.User
	if err := database.Database.Db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		}
		return nil, err
	}

	return &user, nil
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "Successfully Logged out !",
	})
}
