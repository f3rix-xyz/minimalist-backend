package controllers

import (
	"errors"
	"fmt"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v4"
	twilioApi "github.com/twilio/twilio-go/rest/verify/v2"
	"github.com/youruser/yourproject/config"
	"github.com/youruser/yourproject/initializers"
	"github.com/youruser/yourproject/models"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

var secretKey = []byte("secret-key")

func createToken(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":  userID,
			"exp": time.Now().AddDate(0, 1, 0).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		token, err := verifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		idFloat, ok := claims["id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token, id claim missing"})
			c.Abort()
			return
		}
		userID := uint(idFloat)

		// Fetch user details from the database using the id
		var user models.User
		result := initializers.DB.First(&user, userID)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
			c.Abort()
			return
		}

		// Store the user in the context for further use in the handler
		c.Set("user", user)

		c.Next()
	}
}

func ReqOTP(c *gin.Context) {
	var body struct {
		Phone   string `json:"phone" validate:"required,e164"`
		Process string `json:"process" validate:"required"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Process == "signup" {
		log.Println("signup")
		result := initializers.DB.Where("phone = ?", body.Phone).First(&models.User{})
		if result.Error == nil {
			log.Println(result)
			c.JSON(http.StatusConflict, gin.H{"error": "Account already exists. Please log in."})
			return
		}
	}

	params := &twilioApi.CreateVerificationParams{}
	params.SetTo(body.Phone)
	params.SetChannel("sms")

	client, serviceID, err := config.TwilioClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := client.VerifyV2.CreateVerification(serviceID, params)
	if err != nil {
		if strings.Contains(err.Error(), "ApiError 60200") {
			err = errors.New("please check your phone number")
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": "OTP sent successfully", "resp": resp.Status})
}

func CreateUser(c *gin.Context) {
	var body struct {
		Name  string `json:"name" validate:"required"`
		Phone string `json:"phone" validate:"required,e164"`
		OTP   string `json:"otp" validate:"required"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := &twilioApi.CreateVerificationCheckParams{}
	params.SetTo(body.Phone)
	params.SetCode(body.OTP)

	client, serviceID, err := config.TwilioClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := client.VerifyV2.CreateVerificationCheck(serviceID, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if *resp.Status != "approved" {
		c.JSON(400, gin.H{"error": "OTP verification failed", "status": resp.Status})
		return
	} else {
		post := models.User{Name: body.Name, Phone: body.Phone, SubscriptionValidTill: time.Now().AddDate(0, 1, 0)}
		result := initializers.DB.Create(&post)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
			return
		}

		// Create a token
		token, err := createToken(post.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": post, "token": token})
	}
}

func Login(c *gin.Context) {
	var body struct {
		Phone string `json:"phone" validate:"required,e164"`
		OTP   string `json:"otp" validate:"required"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	result := initializers.DB.Where("phone = ?", body.Phone).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	params := &twilioApi.CreateVerificationCheckParams{}
	params.SetTo(body.Phone)
	params.SetCode(body.OTP)

	client, serviceID, err := config.TwilioClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := client.VerifyV2.CreateVerificationCheck(serviceID, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if *resp.Status != "approved" {
		c.JSON(400, gin.H{"error": "OTP verification failed", "status": resp.Status})
		return
	} else {
		// Create a token
		token, err := createToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": "Login successful", "token": token})
	}
}

func Hello(c *gin.Context) {
	c.JSON(200, gin.H{"data": "Hello World"})
}

func Buy(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, gin.H{"data": "You can buy now", "user": user})
}
