package routes

import (
	"encoding/json"
	"fimba/database"
	"fimba/models"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
)

const EARTH_RADIUS float64 = 6371 // km
func IndexPost(c *fiber.Ctx) error {
	// Get the value of the 'id' parameter from the request query
	id := c.Query("id")

	// Get the JWT from the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "No Authorization header provided",
		})
	}
	tokenString := authHeader[len("Bearer "):]

	// Parse the JWT token and validate it
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the JWT secret from the environment variable
		jwtSecret := []byte("your_jwt_secret")
		return jwtSecret, nil
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid JWT token",
		})
	}

	// Check if the 'id' value is valid by querying Redis for it
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer conn.Close()

	isValid, err := redis.Bool(conn.Do("SISMEMBER", "valid_ids", id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if !isValid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid 'id' parameter",
		})
	}

	// Check if the JWT contains the correct user information
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid := uint(claims["uid"].(float64))
		username := claims["uusername"].(string)

		if username != "admin" || uid != 1234 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid credentials",
			})
		}
	} else {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid JWT token",
		})
	}

	authStatus := CheckGpsAuth(id)

	// Send a response with the authentication status
	return c.JSON(fiber.Map{
		"auth_status": authStatus,
	})
}

func CheckGpsAuth(id string) []models.Vehicle {
	var vehicles []models.Vehicle
	db := database.Database.Db
	db.Where("v_api_username = ?", id).Find(&vehicles)
	return vehicles
}

func CheckGeofence(vid uint, lat float64, lng float64, pool *redis.Pool) {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	conn := pool.Get()
	defer conn.Close()

	// Load geofences from Redis
	geofences, err := redis.Values(conn.Do("GEORADIUS", "geofences", lng, lat, 0, "km"))
	if err != nil {
		logger.Fatal(err)
	}

	for _, geo := range geofences {
		geoData := geo.([]interface{})
		geoID := string(geoData[0].([]byte))

		// Check if the vehicle is within the geofence
		geocheck, err := redis.String(conn.Do("GEOPOS", geoID, vid))
		if err != nil {
			logger.Fatal(err)
		}

		// If the vehicle is within the geofence
		if geocheck != "" {
			// Check if an event already exists for this vehicle and geofence today
			geofenceEventKey := fmt.Sprintf("geofence_event:%d:%s:%s", vid, geoID, time.Now().Format("2006-01-02"))
			eventExists, err := redis.Bool(conn.Do("EXISTS", geofenceEventKey))
			if err != nil {
				logger.Fatal(err)
			}

			// If an event doesn't exist, create one
			if !eventExists {
				geoEvent := models.Geofence_events{
					GeVehicleId:  strconv.Itoa(int(vid)),
					GeGeofenceId: geoID,
					Geevent:      "inside",
					Getimestamp:  time.Now().Format("2006-01-02 15:04:05.999999"),
				}
				geoEventJSON, err := json.Marshal(geoEvent)
				if err != nil {
					logger.Fatal(err)
				}

				_, err = conn.Do("SET", geofenceEventKey, geoEventJSON)
				if err != nil {
					logger.Fatal(err)
				}
			}
		}
	}
}

func PositionsPost(c *fiber.Ctx, db *gorm.DB) error {
	t_vehicle := c.FormValue("t_vehicle")
	fromdate := c.FormValue("fromdate")
	todate := c.FormValue("todate")

	if t_vehicle == "" || fromdate == "" || todate == "" {
		return c.JSON(map[string]interface{}{
			"status":  0,
			"message": "Invalid input",
		})
	}

	var positions []models.Positions
	if err := db.Find(&positions, "v_id = ? AND DATE(time) >= ? AND DATE(time) <= ?", t_vehicle, fromdate, todate).Error; err != nil {
		return c.JSON(map[string]interface{}{
			"status":  0,
			"message": err.Error(),
		})
	}

	if len(positions) == 0 {
		return c.JSON(map[string]interface{}{
			"status":  0,
			"message": "No positions found",
		})
	}

	totaldist := TotalDistance(positions[0].ClientLat, positions[0].ClientLng, positions[len(positions)-1].ClientLat, positions[len(positions)-1].ClientLng, EARTH_RADIUS)

	return c.JSON(map[string]interface{}{
		"status":    1,
		"data":      positions,
		"totaldist": totaldist,
		"message":   "data",
	})
}

func HandlePositionsPost(c *fiber.Ctx) error {
	return PositionsPost(c, database.Database.Db)
}

func checkPositions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get vehicle ID and date range from request
	v_id := r.URL.Query().Get("v_id")
	fromdate := r.URL.Query().Get("fromdate")
	todate := r.URL.Query().Get("todate")

	// Query the database for positions within the specified date range
	var positions []models.Positions
	db := database.Database.Db
	db.Where("v_id = ?", v_id).Where("time BETWEEN ? AND ?", fromdate, todate).Find(&positions)

	// Calculate total distance traveled between first and last positions
	distanceFrom := positions[0]
	distanceTo := positions[len(positions)-1]
	totalDist := TotalDistance(distanceFrom.ClientLat, distanceFrom.ClientLng, distanceTo.ClientLat, distanceTo.ClientLng, EARTH_RADIUS)

	// Send response to client
	response := map[string]interface{}{
		"status":    1,
		"data":      positions,
		"totalDist": totalDist,
		"message":   "data",
	}
	json.NewEncoder(w).Encode(response)
}

func TotalDistance(lat1, lon1, lat2, lon2, radius float64) float64 {
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return radius * c
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func CurrentPositionsGet(uname string, gr string, v_id string, db *gorm.DB) (map[string]interface{}, error) {
	data := []map[string]interface{}{}
	positions := []models.Positions{}

	db.Table("positions").Select("positions.*,vehicles.v_name,vehicles.v_type,vehicles.v_color").
		Joins("JOIN vehicles ON vehicles.v_id = positions.v_id").
		Where("vehicles.v_is_active = ?", 1)

	if uname != "" {
		db = db.Where("vehicles.v_api_username = ?", uname)
	}

	if gr != "" {
		db = db.Where("vehicles.v_group = ?", gr)
	}

	if v_id != "" {
		db = db.Where("vehicles.v_id = ?", v_id)
	}

	db.Where("`id` IN (SELECT MAX(id) FROM positions GROUP BY `v_id`)").Find(&positions)

	for _, p := range positions {
		data = append(data, map[string]interface{}{
			"ClientId":  p.ClientId,
			"VehicleId": p.VehicleId,
			"latitude":  p.ClientLat,
			"longitude": p.ClientLng,
			"Status":    p.ClientStatus,
			"altitude":  p.ClientAltitude,
			"speed":     p.ClientSpeed,
			"heading":   p.ClientBearing,
			"accuracy":  p.ClientAccuracy,
			"provider":  p.ClientProvider,
			"timestamp": p.ClientTime.Format("2006-01-02 15:04:05"),
		})
	}

	if len(data) >= 1 {
		return map[string]interface{}{
			"status": 1,
			"data":   data,
		}, nil
	} else {
		return map[string]interface{}{
			"status":  0,
			"message": "No live GPS feed found",
		}, nil
	}
}

func AddPosition(c *fiber.Ctx) error {
	position := new(models.Positions)

	if err := c.BodyParser(position); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  0,
			"message": "Invalid request body",
		})
	}

	if err := database.Database.Db.Create(&position).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  0,
			"message": "Error inserting position into database",
		})
	}

	return c.JSON(fiber.Map{
		"status": 1,
		"data":   position,
	})
}
