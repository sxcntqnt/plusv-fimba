package routes

import (
//	"fimba/models"

	"fmt"
	//"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)
type CustomClaims struct {
    jwt.StandardClaims
    Role string `json:"role"`
    Email    string `json:"email" gorm:"unique"`
    Password string `json:"password"`
    APIKey   string `json:"api_key"`
    LoginId  string `json:"login_id"`

}



func GenerateJWT(c *fiber.Ctx, login *CustomClaims) (string, error) {
    email := login.Email
    password := login.Password
    role := login.Role
    apiKey := login.APIKey
    loginID := login.LoginId

    claims := CustomClaims{
        StandardClaims: jwt.StandardClaims{
            Issuer:    apiKey,
            Subject:   email,
            Audience:  "sxcntcnquntns",
            ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
            IssuedAt:  time.Now().Unix(),
            NotBefore: time.Now().Unix(),
        },
        Role:     role,
        Email:    email,
        Password: password,
        APIKey:   apiKey,
        LoginId:  loginID,
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    secretKey := viper.GetString("jwt.secret_key")
    signedToken, err := token.SignedString([]byte(secretKey))
    if err != nil {
        return "", fmt.Errorf("could not generate JWT token")
    }

    return signedToken, nil
}

func GenerateJWTHandler(c *fiber.Ctx) error {
    var loginData CustomClaims
    if err := c.BodyParser(&loginData); err != nil {
        return err
    }

    // Generate the JWT token using the login data
    token, err := GenerateJWT(c, &loginData)
    if err != nil {
        return err
    }

    // Set the token in a response header or body, depending on your needs
    c.Set("Authorization", "Bearer "+token)

    // Return a success response
    return c.JSON(fiber.Map{
        "token":   token,
        "message": "JWT token generated successfully",
    })
}

func AuthMiddleware(c *fiber.Ctx) error {
	// Get the JWT token from the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Authorization header is missing",
		})
	}
	tokenString := authHeader[len("Bearer "):]

	// Get the secret key from the settings.json file using Viper
	viper.SetConfigName("settings")
	viper.SetConfigType("json")
	viper.AddConfigPath("./config") // Update with the appropriate path to your settings.json file
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	secretKey := viper.GetString("jwt.secret_key")

	// Parse and validate the JWT token using the extracted secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid or expired JWT token",
		})
	}
	// Call the next middleware function or route handler
	return c.Next()
}

