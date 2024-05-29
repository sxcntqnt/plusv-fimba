package database

import (
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

type DbInstance struct {
	Db *gorm.DB
}

var Database DbInstance

func ConnectDb() {
	// Set default values for configuration
	viper.SetDefault("database.filename", "api.db")

	// Read configuration file
	viper.SetConfigName("settings")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("Failed to read configuration file: %s", err)
	}

	// Get database filename from configuration
	dbFilename := "./database/" + viper.GetString("database.filename")

	// Connect to database
	db, err := gorm.Open(sqlite.Open(dbFilename), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database! \n", err.Error())
		os.Exit(2)
	}

	log.Println("Connected to the database successfully")
	db.Logger = logger.Default.LogMode(logger.Info)
	log.Println("running migrations")

	db.AutoMigrate()

	Database = DbInstance{Db: db}
}
