package routes

import (
	"bytes"
	"errors"
	"fimba/database"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Settings struct {
	SettingsId                 uint      `json:"s_id"`
	SettingsCompanyName        string    `json:"s_companyname"`
	SettingsAddress            string    `json:"s_address"`
	SettingsInvoicePrefix      string    `json:"s_inovice_prefix"`
	SettingsLogo               string    `json:"s_logo"`
	SettingsPricePrefix        string    `json:"s_price_prefix"`
	SettingsTermsAndCond       string    `json:"s_inovice_termsandcondition"`
	SettingsInvoiceServiceName string    `json:"s_inovice_servicename"`
	SettingsGoogleApiKey       string    `json:"s_googel_api_key"`
	CreatedAt                  time.Time `gorm:"type:TIMESTAMP(6)"`
}

func UploadFile(c *fiber.Ctx, formFile string, uploadPath string, allowedTypes []string, overwrite bool, maxSize int, maxWidth int, maxHeight int) (string, error) {
	// Retrieve the file from the request body
	file, err := c.FormFile(formFile)
	if err != nil {
		return "", err
	}

	// Check if the file type is allowed
	fileType := filepath.Ext(file.Filename)
	if !isAllowedType(fileType, allowedTypes) {
		return "", errors.New("file type not allowed")
	}

	// Check if the file size is within limits
	if file.Size > int64(maxSize)*1024 {
		return "", errors.New("file size exceeds limit")
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Read the file contents
	fileBytes, err := ioutil.ReadAll(src)
	if err != nil {
		return "", err
	}

	// Check the file dimensions
	imgCfg, _, err := image.DecodeConfig(bytes.NewReader(fileBytes))
	if err != nil {
		return "", err
	}
	if imgCfg.Width > maxWidth || imgCfg.Height > maxHeight {
		return "", errors.New("file dimensions exceed limit")
	}

	// Generate a unique file name
	fileName := uuid.New().String() + fileType

	// Create the destination file
	dst, err := os.Create(filepath.Join(uploadPath, fileName))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Write the file contents to the destination file
	if _, err = dst.Write(fileBytes); err != nil {
		return "", err
	}

	return fileName, nil
}

func isAllowedType(fileType string, allowedTypes []string) bool {
	for _, t := range allowedTypes {
		if t == fileType {
			return true
		}
	}
	return false
}

func GetWebsiteSetting(c *fiber.Ctx) error {
	var s Settings
	err := database.Database.Db.First(&s).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No settings found in the database
			return c.JSON(fiber.Map{})
		}
		// Other error occurred
		return err
	}

	// Return settings
	return c.JSON(s)
}

func WebSettingSave(c *fiber.Ctx) error {
	// Parse JSON request body
	var s Settings
	if err := c.BodyParser(&s); err != nil {
		return err
	}

	// Upload file
	uploadPath := "asset/uploads"
	allowedTypes := []string{"jpg", "jpeg", "png"}
	overwrite := true
	maxSize := int64(1024)
	maxWidth := 50
	maxHeight := 250
	fileName, err := UploadFile(c, "file", uploadPath, allowedTypes, overwrite, int(maxSize), maxWidth, maxHeight)
	if err != nil {
		return err
	}

	// If file was uploaded successfully, update the settings struct with the new file name
	if fileName != "" {
		s.SettingsLogo = fileName
	}

	// Check if settings already exist in the database
	var existingSettings Settings
	result := database.Database.Db.Where("settings_id = ?", s.SettingsId).First(&existingSettings)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// If settings does not exist, create new settings in the database
			result := database.Database.Db.Create(&s)
			if result.Error != nil {
				return result.Error
			}
		} else {
			return result.Error
		}
	} else {
		// If settings already exist, update settings in the database
		result := database.Database.Db.Save(&s)
		if result.Error != nil {
			return result.Error
		}
	}

	// Return success response
	return c.JSON(s)
}

func Logodelete(c *fiber.Ctx) error {
	// Get current settings
	var s Settings
	result := database.Database.Db.First(&s)
	if result.Error != nil {
		return result.Error
	}

	// Delete the logo file
	err := os.Remove(s.SettingsLogo)
	if err != nil {
		return err
	}

	// Update the settings with a blank logo field
	s.SettingsLogo = ""
	result = database.Database.Db.Save(&s)
	if result.Error != nil {
		return result.Error
	}

	// Return success response
	return c.SendStatus(fiber.StatusOK)
}
