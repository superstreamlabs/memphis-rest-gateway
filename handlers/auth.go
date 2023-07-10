/*
	Memphis authentication.
*/

package handlers

import (
	"rest-gateway/conf"
	"rest-gateway/logger"
	"rest-gateway/models"
	"rest-gateway/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/memphisdev/memphis.go"
)

var configuration = conf.GetConfig()

type AuthHandler struct{}

func (ah AuthHandler) Authenticate(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	var body models.AuthSchema

	stationName := c.Params("stationName")

	user_pass_based_auth := configuration.USER_PASS_BASED_AUTH
	client_cert_path := configuration.CLIENT_CERT_PATH
	client_key_path := configuration.CLIENT_KEY_PATH
	root_ca_path := configuration.ROOT_CA_PATH
	memphis_host := configuration.MEMPHIS_HOST
	jwt_expires_in_minutes := configuration.JWT_EXPIRES_IN_MINUTES
	jwt_secret := configuration.JWT_SECRET
	refresh_jwt_secret := configuration.REFRESH_JWT_SECRET

	for _, station := range configuration.Stations {
		if station.NAME == stationName {
			if station.JWT_EXPIRES_IN_MINUTES != 0 {
				jwt_expires_in_minutes = station.JWT_EXPIRES_IN_MINUTES
			}

			if station.JWT_SECRET != "" {
				jwt_secret = station.JWT_SECRET
			}

			if station.REFRESH_JWT_SECRET != "" {
				refresh_jwt_secret = station.REFRESH_JWT_SECRET
			}
		}
	}

	if err := c.BodyParser(&body); err != nil {
		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}

	var conn *memphis.Conn
	var err error
	opts := []memphis.Option{memphis.Reconnect(true), memphis.MaxReconnect(10), memphis.ReconnectInterval(3 * time.Second)}
	if user_pass_based_auth {
		opts = append(opts, memphis.Password(body.Password))
	} else {
		opts = append(opts, memphis.ConnectionToken(body.ConnectionToken))
	}
	if client_cert_path != "" && client_key_path != "" && root_ca_path != "" {
		opts = append(opts, memphis.Tls(client_cert_path, client_key_path, root_ca_path))
	}
	conn, err = memphis.Connect(memphis_host, body.Username, opts...)

	if err != nil {
		if strings.Contains(err.Error(), "Authorization Violation") || strings.Contains(err.Error(), "token") {
			log.Warnf("Authentication error")
			return c.Status(401).JSON(fiber.Map{
				"message": "Wrong credentials",
			})
		}

		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}
	conn.Close()
	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins, jwt_expires_in_minutes, jwt_secret, refresh_jwt_secret)
	if err != nil {
		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Create tokens error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":                      token,
		"expires_in":               tokenExpiry * 60 * 1000,
		"jwt_refresh_token":        refreshToken,
		"refresh_token_expires_in": refreshTokenExpiry * 60 * 1000,
	})
}

func createTokens(tokenExpiryMins int, refreshTokenExpiryMins int, jwt_expires_in_minutes int, jwt_secret string, refresh_jwt_secret string) (string, string, int, int, error) {
	if tokenExpiryMins <= 0 {
		tokenExpiryMins = jwt_expires_in_minutes
	}

	if refreshTokenExpiryMins <= 0 {
		refreshTokenExpiryMins = jwt_expires_in_minutes
	}

	atClaims := jwt.MapClaims{}
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(tokenExpiryMins)).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(jwt_secret))
	if err != nil {
		return "", "", 0, 0, err
	}

	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(refreshTokenExpiryMins)).Unix()
	at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	refreshToken, err := at.SignedString([]byte(refresh_jwt_secret))
	if err != nil {
		return "", "", 0, 0, err
	}
	return token, refreshToken, tokenExpiryMins, refreshTokenExpiryMins, nil
}

func (ah AuthHandler) RefreshToken(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	var body models.RefreshTokenSchema

	stationName := c.Params("stationName")

	jwt_expires_in_minutes := configuration.JWT_EXPIRES_IN_MINUTES
	jwt_secret := configuration.JWT_SECRET
	refresh_jwt_secret := configuration.REFRESH_JWT_SECRET

	for _, station := range configuration.Stations {
		if station.NAME == stationName {
			if station.JWT_EXPIRES_IN_MINUTES != 0 {
				jwt_expires_in_minutes = station.JWT_EXPIRES_IN_MINUTES
			}

			if station.JWT_SECRET != "" {
				jwt_secret = station.JWT_SECRET
			}

			if station.REFRESH_JWT_SECRET != "" {
				refresh_jwt_secret = station.REFRESH_JWT_SECRET
			}
		}
	}

	if err := c.BodyParser(&body); err != nil {
		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}

	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins, jwt_expires_in_minutes, jwt_secret, refresh_jwt_secret)
	if err != nil {
		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Create tokens error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":                      token,
		"expires_in":               tokenExpiry * 60 * 1000,
		"jwt_refresh_token":        refreshToken,
		"refresh_token_expires_in": refreshTokenExpiry * 60 * 1000,
	})
}
