package handlers

import (
	"fmt"
	"rest-gateway/conf"
	"rest-gateway/logger"
	"rest-gateway/models"
	"rest-gateway/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/memphisdev/memphis.go"
)

var configuration = conf.GetConfig()

type AuthHandler struct{}

type Connection struct {
	Connection *memphis.Conn `json:"connection"`
	ExpireDate float64       `json:"expire_date"`
}

var connectionsCache = map[string]map[string]Connection{}

func connect(password, username, connectionToken string, accountId int) (*memphis.Conn, error) {
	var err error
	if accountId == 0 {
		accountId = 1
	}
	opts := []memphis.Option{memphis.Reconnect(true), memphis.MaxReconnect(10), memphis.ReconnectInterval(3 * time.Second)}
	if configuration.USER_PASS_BASED_AUTH {
		opts = append(opts, memphis.Password(password), memphis.AccountId(accountId))
	} else {
		opts = append(opts, memphis.ConnectionToken(connectionToken))
	}
	if configuration.CLIENT_CERT_PATH != "" && configuration.CLIENT_KEY_PATH != "" && configuration.ROOT_CA_PATH != "" {
		opts = append(opts, memphis.Tls(configuration.CLIENT_CERT_PATH, configuration.CLIENT_KEY_PATH, configuration.ROOT_CA_PATH))
	}
	conn, err := memphis.Connect(configuration.MEMPHIS_HOST, username, opts...)
	if err != nil {
		return conn, err
	}
	return conn, nil
}

func (ah AuthHandler) Authenticate(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	var body models.AuthSchema
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

	if body.AccountId == 0 || !configuration.USER_PASS_BASED_AUTH {
		body.AccountId = 1
	}

	conn, err := connect(body.Password, body.Username, body.ConnectionToken, body.AccountId)
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

	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins, body.Username, body.AccountId, body.Password, body.ConnectionToken)
	if err != nil {
		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Create tokens error",
		})
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(configuration.JWT_SECRET), nil
	})
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		log.Errorf("Authenticate: Claims are not of the expected type")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Claims are not of the expected type",
		})
	}

	tenantName := strconv.Itoa(int(claims["account_id"].(float64)))
	username := claims["username"].(string)
	tokenExpire := claims["exp"].(float64)

	if connectionsCache[tenantName] == nil {
		connectionsCache[tenantName] = make(map[string]Connection)
	}

	connectionsCache[tenantName][username] = Connection{Connection: conn, ExpireDate: tokenExpire}
	// conn.Close()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":                      token,
		"expires_in":               tokenExpiry * 60 * 1000,
		"jwt_refresh_token":        refreshToken,
		"refresh_token_expires_in": refreshTokenExpiry * 60 * 1000,
	})
}

func createTokens(tokenExpiryMins, refreshTokenExpiryMins int, username string, accountId int, password, connectionToken string) (string, string, int, int, error) {
	if tokenExpiryMins <= 0 {
		tokenExpiryMins = configuration.JWT_EXPIRES_IN_MINUTES
	}

	if refreshTokenExpiryMins <= 0 {
		refreshTokenExpiryMins = configuration.JWT_EXPIRES_IN_MINUTES
	}

	atClaims := jwt.MapClaims{}
	atClaims["username"] = username
	if accountId == 0 || !configuration.USER_PASS_BASED_AUTH {
		accountId = 1
	}
	atClaims["account_id"] = accountId
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(tokenExpiryMins)).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(configuration.JWT_SECRET))
	if err != nil {
		return "", "", 0, 0, err
	}

	atClaims["password"] = password
	atClaims["connection_token"] = connectionToken
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(refreshTokenExpiryMins)).Unix()
	at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	refreshToken, err := at.SignedString([]byte(configuration.REFRESH_JWT_SECRET))
	if err != nil {
		return "", "", 0, 0, err
	}
	return token, refreshToken, tokenExpiryMins, refreshTokenExpiryMins, nil
}

func (ah AuthHandler) RefreshToken(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	var body models.RefreshTokenSchema
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

	parsedRefreshToken, err := jwt.Parse(body.JwtRefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(configuration.REFRESH_JWT_SECRET), nil
	})
	claims, ok := parsedRefreshToken.Claims.(jwt.MapClaims)
	if !ok {
		log.Errorf("RefreshToken: Claims are not of the expected type")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Claims are not of the expected type",
		})
	}

	tenantName := strconv.Itoa(int(claims["account_id"].(float64)))
	username := claims["username"].(string)
	tokenExpire := claims["exp"].(float64)
	password := claims["password"].(string)
	connectionToken := claims["connection_token"].(string)
	//

	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins, username, int(claims["account_id"].(float64)), password, connectionToken)
	if err != nil {
		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Create tokens error",
		})
	}

	conn, err := connect(password, username, connectionToken, int(claims["account_id"].(float64)))
	if err != nil {
		if strings.Contains(err.Error(), "Authorization Violation") || strings.Contains(err.Error(), "token") {
			log.Warnf("Authentication error")
			return c.Status(401).JSON(fiber.Map{
				"message": "Wrong credentials",
			})
		}

		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}

	if connectionsCache[tenantName] == nil {
		connectionsCache[tenantName] = make(map[string]Connection)
	}

	connectionsCache[tenantName][username] = Connection{Connection: conn, ExpireDate: tokenExpire}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":                      token,
		"expires_in":               tokenExpiry * 60 * 1000,
		"jwt_refresh_token":        refreshToken,
		"refresh_token_expires_in": refreshTokenExpiry * 60 * 1000,
	})
}

func CleanConnectionsCache() {
	for range time.Tick(time.Second * 30) {
		for t, tenant := range connectionsCache {
			for u, user := range tenant {
				fmt.Println(user)

				currentTime := time.Now()
				unixTimeNow := currentTime.Unix()

				conn := connectionsCache[t][u].Connection
				conn.Close()

				if unixTimeNow > int64(user.ExpireDate) {
					delete(connectionsCache[t], u)
				}
			}
			if len(connectionsCache[t]) == 0 {
				delete(connectionsCache, t)
			}
		}
	}
}
