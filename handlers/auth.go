package handlers

import (
	"http-proxy/conf"
	"http-proxy/models"
	"http-proxy/utils"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/memphisdev/memphis.go"
)

var configuration = conf.GetConfig()

type AuthHandler struct{}

func (ah AuthHandler) Authenticate(c *fiber.Ctx) error {
	var body models.AuthSchema
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}
	conn, err := memphis.Connect(configuration.MEMPHIS_HOST, body.Username, body.ConnectionToken)
	if err != nil {
		if strings.Contains(err.Error(), "Authorization Violation") {
			return c.Status(401).JSON(fiber.Map{
				"message": "Wrong credentials",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}
	conn.Close()
	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Create tokens error",
		})
	}
	cookie := new(fiber.Cookie)
	cookie.Name = "jwt-refresh-token"
	cookie.Value = refreshToken
	cookie.MaxAge = refreshTokenExpiry * 60 * 1000
	cookie.Path = "/"
	cookie.Domain = ""
	cookie.Secure = false
	cookie.HTTPOnly = true
	c.Cookie(cookie)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":        token,
		"expires_in": tokenExpiry * 60 * 1000,
	})
}

func createTokens(tokenExpiryMins, refreshTokenExpiryMins int) (string, string, int, int, error) {
	if tokenExpiryMins <= 0 {
		tokenExpiryMins = configuration.JWT_EXPIRES_IN_MINUTES
	}

	if refreshTokenExpiryMins <= 0 {
		refreshTokenExpiryMins = configuration.JWT_EXPIRES_IN_MINUTES
	}

	atClaims := jwt.MapClaims{}
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(tokenExpiryMins)).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(configuration.JWT_SECRET))
	if err != nil {
		return "", "", 0, 0, err
	}

	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(refreshTokenExpiryMins)).Unix()
	at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	refreshToken, err := at.SignedString([]byte(configuration.REFRESH_JWT_SECRET))
	if err != nil {
		return "", "", 0, 0, err
	}
	return token, refreshToken, tokenExpiryMins, refreshTokenExpiryMins, nil
}

func (ah AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var body models.RefreshTokenSchema
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}

	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Create tokens error",
		})
	}
	cookie := new(fiber.Cookie)
	cookie.Name = "jwt-refresh-token"
	cookie.Value = refreshToken
	cookie.MaxAge = refreshTokenExpiry * 60 * 1000
	cookie.Path = "/"
	cookie.Domain = ""
	cookie.Secure = false
	cookie.HTTPOnly = true
	c.Cookie(cookie)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":        token,
		"expires_in": tokenExpiry * 60 * 1000,
	})
}
