package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/xjem/t38c"

	"fimba/database"
	"fimba/models"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Positions struct {
	ClientId       uint      `json:"id"`
	ClientTime     time.Time `gorm:"type:TIMESTAMP(6)"`
	VehicleId      int       `json:"v_id"`
	ClientLat      float64   `json:"latitude"`
	ClientLng      float64   `json:"longitude"`
	ClientStatus   string    `json:"Raining/Clear"`
	ClientAltitude float64   `json:"altitude"`
	ClientSpeed    float64   `json:"speed"`
	ClientBearing  float64   `json:"bearing"`
	ClientAccuracy int       `json:"accuracy"`
	ClientProvider string    `json:"provider"`
	ClientComment  string    `json:"comment"`
	CreatedAt      time.Time `gorm:"type:TIMESTAMP(6)"`
}

type LocationHistory struct {
	Lat     float64   `json:"latitude"`
	Lon     float64   `json:"longitude"`
	Created time.Time `json:"created_at"`
}

func config() error {
	client, err := t38c.New(t38c.Config{
		Address: "localhost:9851",
		Debug:   true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	return nil
}

func TrackVehicle(c *fiber.Ctx) error {
	// parse vehicle data from request
	vehicle := models.Positions{}
	if err := c.BodyParser(&vehicle); err != nil {
		return err
	}

	// save vehicle data to database
	if err := database.Database.Db.Create(&vehicle).Error; err != nil {
		return err
	}

	// create a redis connection pool
	pool := &redis.Pool{
		MaxIdle:   10,
		MaxActive: 100,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:9851")
		},
	}

	// get a redis connection from the pool
	conn := pool.Get()
	defer conn.Close()

	// save vehicle location to Tile38
	args := []interface{}{"vehicles", strconv.Itoa(int(vehicle.VehicleId)), "POINT", strconv.FormatFloat(vehicle.ClientLat, 'f', -1, 64), strconv.FormatFloat(vehicle.ClientLng, 'f', -1, 64)}
	_, err := conn.Do("SET", args...)
	if err != nil {
		return err
	}

	// return success response
	return c.JSON(fiber.Map{
		"message": "Vehicle tracked successfully",
		"vehicle": vehicle,
	})
}
func GetLocationHistory(c *fiber.Ctx) error {
	// get vehicle ID from request parameters
	id := c.Params("id")

	// parse query parameters for start and end dates
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	var startDate, endDate time.Time
	var err error
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return err
		}
	}
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return err
		}
		// add 24 hours to end date to include all points on the end date
		endDate = endDate.Add(24 * time.Hour)
	} else {
		// if end date is not provided, default to now
		endDate = time.Now()
	}

	// create Redis connection pool
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}
	defer pool.Close()

	// get Redis connection from pool
	conn := pool.Get()
	defer conn.Close()

	// query location history from Redis
	var cursor int
	var keys []string
	for {
		arr, err := redis.Values(conn.Do("SCAN", cursor, "MATCH", id+"*", "COUNT", 1000))
		if err != nil {
			return err
		}
		cursor, _ = redis.Int(arr[0], nil)
		ks, _ := redis.Strings(arr[1], nil)
		keys = append(keys, ks...)
		if cursor == 0 {
			break
		}
	}

	// sort keys by timestamp
	sort.Slice(keys, func(i, j int) bool {
		ti, _ := strconv.Atoi(strings.TrimPrefix(keys[i], id+"_"))
		tj, _ := strconv.Atoi(strings.TrimPrefix(keys[j], id+"_"))
		return ti < tj
	})

	// parse location history into a slice of LocationHistory structs
	var history []LocationHistory
	for _, key := range keys {
		loc, err := redis.Values(conn.Do("LRANGE", key, "0", "-1"))
		if err != nil {
			return err
		}
		lat, err := redis.Float64(loc[1], nil)
		if err != nil {
			return err
		}
		lon, err := redis.Float64(loc[0], nil)
		if err != nil {
			return err
		}
		ts, _ := time.Parse(time.RFC3339, strings.TrimPrefix(key, id+"_"))
		// filter location points by date range
		if (startDate.IsZero() || ts.After(startDate)) && ts.Before(endDate) {
			history = append(history, LocationHistory{
				Lat:     lat,
				Lon:     lon,
				Created: ts,
			})
		}
	}

	// return location history as JSON response
	return c.JSON(history)
}
