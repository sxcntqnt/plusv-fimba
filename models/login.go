package models

type LoginData struct {
	Number   string `json:"number" gorm "unique"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
	Role     string `json:"role"`
        APIKey   string `json:"api_key"`
    LoginId  string `json:"login_id"`
}
