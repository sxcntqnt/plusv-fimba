package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fimba/database"
	"fimba/models"
	"fimba/routes"
	"github.com/gofiber/fiber/v2"
	mailslurp "github.com/mailslurp/mailslurp-client-go"
	"github.com/spf13/viper"
)

var inbox *mailslurp.InboxDto
var smtpSettings models.SettingsSMTP

func welcome(c *fiber.Ctx) error {
	return c.SendString("Welcome to sxcntcnquntns")
}

func setupRoutes(app *fiber.App) {
	//api routes
	app.Get("/api", welcome)
	app.Get("/api/indexPost", routes.IndexPost)
	app.Post("/api/positions", routes.HandlePositionsPost)
	app.Put("/api/addPositions", routes.AddPosition)

	// Users endpoints
	app.Post("/api/Reguser", routes.RegisterUser)
	app.Get("/api/Logoutuser", routes.Logout)
	app.Get("/api/users/:id", routes.CurrentUser)
	app.Put("/api/users/:id", routes.UpdateUser)
	app.Delete("/api/Deluser", routes.DeleteUser)

	// Vehicle endpoints
	app.Post("/api/CreateVehicle", routes.CreateVehicle)
	app.Get("/api/GetvehicleAll", routes.GetVehicle)
	app.Get("/api/Getvehicles/:id", routes.GetVehicles)
	app.Put("/api/Updatevehicle/:id", routes.UpdateVehicle)
	app.Delete("/api/Delvehicles", routes.DeleteVehicle)
	app.Post("/api/CreateVehicles-group", routes.CreateVehicleGroup)
	app.Post("/api/Addvehicles-group", routes.AddVehiclesToGroup)
	app.Get("/api/Getvehicle-group", routes.GetVehicleGroup)
	app.Get("/api/Getvehicle-group/:name", routes.GetVehicleGroupByName)
	app.Put("api/Updatevehicle-group/:id", routes.UpdateVehicleGroup)
	app.Delete("/api/Delvehicle-groups/:id", routes.DeleteVehiclesFromGroup)

	// Driver endpoints
	app.Post("/api/CreateDriver", routes.CreateDriver)
	app.Get("/api/GetDriverAll", routes.GetDriver)
	app.Get("/api/GetDriver/:id", routes.GetDriver)
	app.Put("/api/UpdateDriver/:id", routes.UpdateDriver)
	app.Delete("/api/DelDriver", routes.DeleteDriver)

	// Booking endpoints
	app.Get("/api/Getbookings/:id", routes.GetBooking)
	app.Post("/api/BookVehicle", routes.BookVehicle)
	app.Get("/api/EditBooking", routes.EditBooking)
	app.Get("/api/ListBookings", routes.ListBookings)

	// Customer endpoints
	app.Post("/api/CreateCustomer", routes.CreateCustomer)
	app.Get("/api/GetCustomerAll", routes.GetCustomer)
	app.Get("/api/GetCustomer/:id", routes.GetCustomer)
	app.Put("/api/UpdateCustomer/:id", routes.UpdateCustomer)
	app.Delete("/api/DelCustomer", routes.DeleteCustomer)

	// Fuel endpoints
	app.Post("/api/CreateFuel", routes.CreateFuelEntry)
	app.Get("/api/EditFuel", routes.EditFuelEntry)
	app.Get("/api/ListFuel/:id", routes.ListFuelEntries)
	// Register a POST route for adding to expenses
	app.Post("/api/expense", func(c *fiber.Ctx) error {
		// Call the AddToExpense function
		err := routes.AddToExpense(c, nil) // Update with any required argument for AddToExpense
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Send a response indicating that the fuel entry was added to expenses successfully
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Fuel entry added to expenses successfully"})
	})

	//Reminder endpoints
	app.Post("api/CreateReminder", routes.CreateReminder)
	app.Get("api/GetReminder/:id", routes.GetReminder)
	app.Put("api/EditReminder/:id", routes.EditReminder)
	app.Get("api/Listreminders", routes.ListReminders)

	// Income and Expenses Endpoint
	app.Post("api/AddIncomexpense", routes.AddIncExp)
	app.Get("api/GetIncomexpense", routes.GetIncomes)
	app.Get("api/GetIncomexpense/:id", routes.GetIncome)
	app.Put("api/EditIncomexpense/:id", routes.EditIncome)

	// Tracking and Live location Endpoints
	app.Post("api/tracking/vehicles", routes.TrackVehicle)
	app.Get("api/tracking/:id/history", routes.GetLocationHistory)

	// Geofence endpoints
	app.Post("api/CreateGeofences", routes.CreateGeofence)
	app.Get("api/VehicleGeofence/:id", routes.VehicleEnterGeofence)
	app.Get("api/GetGeoEvents/:id", routes.GetGeoEvents)
	app.Get("api/GetVehicleGeofence/", routes.GetVehicleForGeofence)
	app.Delete("api/DelGeofence", routes.DeleteGeofence)

	// Reports endpoints {booking report, income and expense report, fuel report}
	app.Post("api/reports/incomeexpense", routes.GenerateIncExpReport)
	app.Get("api/reports/booking", routes.GenerateBookingReport)
	app.Get("api/reports/fuel/:id", routes.GenerateFuelReport)

	// General Settings endpoint
	app.Post("api/settings/upload", func(c *fiber.Ctx) error {
		formFile, err := c.FormFile("file")
		if err != nil {
			// Handle error
		}
		filename := formFile.Filename
		uploadPath := "./uploads"
		allowedTypes := []string{"jpg", "png"}
		overwrite := false
		maxSize := 1024 * 1024 * 5
		maxWidth := 1024
		maxHeight := 1024
		_, err = routes.UploadFile(c, filename, uploadPath, allowedTypes, overwrite, maxSize, maxWidth, maxHeight)
		if err != nil {
			// Handle error
		}
		return nil
	})

	app.Get("api/settings/WebSetting", routes.GetWebsiteSetting)
	app.Post("api/settings/WebSettingSave", routes.WebSettingSave)
	app.Delete("api/settings/LogoDel", routes.Logodelete)

	//mail endpoints
	app.Post("api/mail/SmptConfSave", routes.Smtpconfigsave)
	app.Post("api/mail/emailtemplate", routes.EmailTemplate)
	app.Put("api/mail/emailtemplate/:id", routes.UpdateEmailTemplate)

	// Email test endpoint
	app.Post("/api/mail/email-test", func(c *fiber.Ctx) error {
		// Load SMTP settings from configuration file
		absPath, err := os.Stat("./config/settings.json")
		if err != nil {
			log.Fatalf("Failed to get absolute path of configuration file: %v", err)
		}
		viper.SetConfigFile(absPath.Name())
		err = viper.ReadInConfig()
		if err != nil {
			log.Fatalf("Failed to load configuration file: %v", err)
		}

		// Get SMTP settings from configuration file
		smtpSettings := viper.GetStringMap("smtp")

		// Get inbox email address from SMTP settings
		inboxEmailAddress := smtpSettings["smtp_emailfrom"].(string)

		// Return success message
		return c.JSON(fiber.Map{
			"message": "Inbox created for testing.",
			"address": inboxEmailAddress,
		})
	})

	// Password endpoint
	app.Post("api/users/:id/reset-password", routes.ResetPassword)

	// Dashboard endpoint
	app.Put("api/dashboard", routes.DashboardHandler)

}
func main() {
	absPath, err := filepath.Abs("./config/settings.json")
	if err != nil {
		log.Fatalf("Failed to get absolute path of configuration file: %v", err)
	}

	// Read configuration file
	viper.SetConfigFile(absPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	viper.AddConfigPath("config/")

	var config models.SettingsSMTP

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Failed to unmarshal SMTP settings from configuration file: %v", err)
	}

	// Initialize MailSlurp client and inbox
	apiKey := viper.GetString("email.apiKey")
	client, inbox, err := routes.InitMailSlurpClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to initialize MailSlurp client: %v", err)
	}

	ctx := &fiber.Ctx{} // create a new fiber.Ctx instanc
	// Test SMTP settings
	err = routes.SMTPConfigTestEmail(ctx, apiKey, inbox, client, smtpSettings)
	if err != nil {
		fmt.Printf("Could not fetch settings: %v\n", err)
	}

	// Setup routes and start server
	database.ConnectDb()
	app := fiber.New()

	// Apply the AuthMiddleware to all routes except "/api/GenJwt"
	app.Use(func(c *fiber.Ctx) error {
		if c.Path() != "/api/GenJwt" {
			return routes.AuthMiddleware(c)
		}
		return c.Next()
	})

	// Define the route that doesn't require authentication
	app.Post("/api/GenJwt", routes.GenerateJWTHandler)

	go setupRoutes(app)
	log.Fatal(app.Listen(":3420"))
}
