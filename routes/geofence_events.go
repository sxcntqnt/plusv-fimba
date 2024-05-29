package routes

import (
	"encoding/json"
	"fimba/database"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gomodule/redigo/redis"
)

type Geofence struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name" gorm:"column:name"`
	Latitude  float64   `json:"latitude" gorm:"column:latitude"`
	Longitude float64   `json:"longitude" gorm:"column:longitude"`
	Radius    float64   `json:"radius" gorm:"column:radius"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

type geofence_events struct {
	Geo_id        uint      `json:"ge_id"`
	Ge_VehicleId  string    `json:"ge_v_id"`
	Ge_GeofenceId string    `json:"ge_geo_id"`
	Ge_event      string    `json:"ge_event"`
	Ge_timestamp  string    `json:"type:timestamp(6)"`
	CreatedAt     time.Time `gorm:"type:TIMESTAMP(6)"`
}

func CreateGeofence(c *fiber.Ctx) error {
	// parse geofence data from request body
	var geofence Geofence
	if err := c.BodyParser(&geofence); err != nil {
		return err
	}

	// create Redis connection pool
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", ":6379")
		},
	}

	// create Redis connection from pool
	conn := pool.Get()
	defer conn.Close()

	// build Tile38 command to set a geofence
	cmd := fmt.Sprintf("SET %s POINT %f %f %f", geofence.Name, geofence.Longitude, geofence.Latitude, geofence.Radius)

	// send command to Tile38 using Redis connection
	_, err := conn.Do("GEOFENCE", cmd)
	if err != nil {
		return err
	}

	// return success response
	return c.JSON(fiber.Map{
		"message": "Geofence created successfully",
		"name":    geofence.Name,
	})

}
func SaveGeofenceEvent(geofenceId, vehicleId, eventType string) error {
	// create Redis connection
	conn, err := redis.Dial("tcp", ":9851")
	if err != nil {
		return err
	}
	defer conn.Close()

	// save geofence event to Redis
	_, err = conn.Do("RPUSH", "geofence_events", geofenceId, vehicleId, eventType)
	if err != nil {
		return err
	}

	return nil
}

func VehicleEnterGeofence(c *fiber.Ctx) error {
	// parse geofence data from request body
	var geofence Geofence
	if err := c.BodyParser(&geofence); err != nil {
		return err
	}

	// create Redis connection
	conn, err := redis.Dial("tcp", ":9851")
	if err != nil {
		return err
	}
	defer conn.Close()

	// check if vehicle is within geofence
	result, err := redis.String(conn.Do("GEORADIUS", "vehicles", geofence.Longitude, geofence.Latitude, geofence.Radius, "m", "WITHCOORD", "COUNT", "1"))
	if err != nil {
		return err
	}

	// if result is not empty, save geofence event to Redis
	if result != "" {
		if err := SaveGeofenceEvent(geofence.Name, result, "enter"); err != nil {
			return err
		}
	}

	return nil
}
func SaveGeofence(c *fiber.Ctx) error {
	// parse geofence data from request body
	var geofence Geofence
	if err := c.BodyParser(&geofence); err != nil {
		return err
	}

	// create Redis connection pool
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", ":9851")
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	defer pool.Close()

	// add geofence to Tile38
	conn := pool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", "geofences", geofence.Name, "POINT", geofence.Latitude, geofence.Longitude, "OBJECT", geofence.Name, "WITHIN", geofence.Radius)
	if err != nil {
		return err
	}

	// create geofence record in database
	database.Database.Db.Create(&geofence)

	// save geofence event to Redis
	conn = pool.Get()
	defer conn.Close()
	_, err = conn.Do("RPUSH", "geofence_events", geofence.ID, "", "created")
	if err != nil {
		return err
	}

	// return success response
	return c.JSON(fiber.Map{
		"message":  "Geofence saved successfully",
		"geofence": geofence,
	})
}

func DeleteGeofence(c *fiber.Ctx) error {
	// get geofence name from URL parameter
	name := c.Params("name")

	// delete geofence record from database
	result := database.Database.Db.Delete(&Geofence{}, "name = ?", name)
	if result.Error != nil {
		return result.Error
	}

	// create Redis connection
	conn, err := redis.Dial("tcp", ":9851")
	if err != nil {
		return err
	}
	defer conn.Close()

	// delete geofence from Tile38
	_, err = conn.Do("DELHOOK", name)
	if err != nil {
		return err
	}

	// return success response
	return c.JSON(fiber.Map{
		"message": "Geofence deleted successfully",
		"name":    name,
	})
}
func GetGeoEvents(c *fiber.Ctx) error {
	// Get the start and end dates from the query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Parse the start and end dates
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start date format",
		})
	}
	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end date format",
		})
	}

	// Get the geofence name from the request URL
	name := c.Params("name")

	// Create a Redis client and get a connection
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return err
	}
	defer conn.Close()

	// Query Redis for geofence events within the time range for the given geofence name
	geofenceEvents, err := redis.Strings(conn.Do("ZRANGEBYSCORE", name, startDate.Unix(), endDate.Unix()))
	if err != nil {
		return err
	}

	// Parse the geofence events into a slice of struct objects
	var events []geofence_events
	for _, event := range geofenceEvents {
		var ge geofence_events
		if err := json.Unmarshal([]byte(event), &ge); err != nil {
			return err
		}
		events = append(events, ge)
	}

	// Return the list of events as a JSON response
	return c.JSON(events)
}
func GetVehicleForGeofence(c *fiber.Ctx) error {
	// Get the geofence name from the request URL
	geofenceName := c.Params("name")

	// Query the database for the vehicle associated with the given geofence name
	var vehicle Vehicle
	if err := database.Database.Db.Joins("GeofenceVehicles").Where("geofence_vehicles.geofence_id IN (SELECT id FROM geofences WHERE name = ?)", geofenceName).First(&vehicle).Error; err != nil {
		return err
	}

	// Return the vehicle name as a JSON response
	return c.JSON(fiber.Map{
		"vehicle_name": vehicle.V_Name,
	})
}
