package handlers

import (
	"fmt"
	"rest-gateway/conf"
	"rest-gateway/logger"
	"rest-gateway/models"
	"rest-gateway/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/memphisdev/memphis.go"
)

var configuration = conf.GetConfig()
var lock sync.Mutex

type AuthHandler struct{}

type Connection struct {
	Connection     *memphis.Conn `json:"connection"`
	ExpirationTime int64         `json:"expiration_date"`
}

var connectionsCache = map[string]map[string]Connection{}

func connect(password, username, connectionToken string, accountId int) (*memphis.Conn, error) {
	var err error
	opts := []memphis.Option{memphis.Reconnect(true), memphis.MaxReconnect(10), memphis.ReconnectInterval(3 * time.Second)}
	if configuration.USER_PASS_BASED_AUTH {
		if accountId == 0 {
			accountId = 1
		}
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
			"message": "Server error",
		})
	}

	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}

	conn, err := connect(body.Password, body.Username, body.ConnectionToken, int(body.AccountId))
	if err != nil {
		if strings.Contains(err.Error(), "Authorization Violation") || strings.Contains(err.Error(), "token") {
			log.Warnf("Authentication error")
			return c.Status(401).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}

		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}

	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins, body.Username, int(body.AccountId), body.Password, body.ConnectionToken)
	if err != nil {
		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}

	tokenExpiration := time.Now().Add(time.Minute * time.Duration(body.TokenExpiryMins)).Unix()
	username := strings.ToLower(body.Username)
	accountId := strconv.Itoa(int(body.AccountId))
	if connectionsCache[accountId] == nil {
		lock.Lock()
		connectionsCache[accountId] = make(map[string]Connection)
		lock.Unlock()
	}

	lock.Lock()
	connectionsCache[accountId][username] = Connection{Connection: conn, ExpirationTime: tokenExpiration}
	lock.Unlock()

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
	if configuration.USER_PASS_BASED_AUTH {
		if accountId == 0 {
			accountId = 1
		}
	}
	atClaims["username"] = username
	atClaims["password"] = password
	atClaims["account_id"] = accountId
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(tokenExpiryMins)).Unix()
	atClaims["connection_token"] = connectionToken
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
	log := logger.GetLogger(c)
	var body models.RefreshTokenSchema
	if err := c.BodyParser(&body); err != nil {
		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}
	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}
	userData, ok := c.Locals("userData").(models.AuthSchema)
	if !ok {
		log.Errorf("CreateHandleMessage - handleHeaders:  failed get the user data from middleare")
		c.Status(500)
		return c.JSON(&fiber.Map{
			"success": false,
			"error":   "Server error",
		})
	}

	username := userData.Username
	accountId := int(userData.AccountId)
	password := userData.Password
	connectionToken := userData.ConnectionToken

	conn, err := connect(password, username, connectionToken, accountId)
	if err != nil {
		if strings.Contains(err.Error(), "Authorization Violation") || strings.Contains(err.Error(), "token") {
			log.Warnf("RefreshToken: Authentication error")
			return c.Status(401).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}

		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}

	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins, username, accountId, password, connectionToken)
	if err != nil {
		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}
	tokenExpiration := time.Now().Add(time.Minute * time.Duration(body.TokenExpiryMins)).Unix()

	accountId = int(accountId)
	if connectionsCache[strconv.Itoa(int(accountId))] == nil {
		lock.Lock()
		connectionsCache[strconv.Itoa(accountId)] = make(map[string]Connection)
		lock.Unlock()
	}

	lock.Lock()
	connectionsCache[strconv.Itoa(accountId)][username] = Connection{Connection: conn, ExpirationTime: tokenExpiration}
	lock.Unlock()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":                      token,
		"expires_in":               tokenExpiry * 60 * 1000,
		"jwt_refresh_token":        refreshToken,
		"refresh_token_expires_in": refreshTokenExpiry * 60 * 1000,
	})
}

func CleanConnectionsCache() {
	for range time.Tick(time.Second * 30) {
		connectionsCache := map[string]map[string]Connection{}
		fmt.Println("connectionsCache", connectionsCache)
		for t, tenant := range connectionsCache {
			for u, user := range tenant {
				currentTime := time.Now()
				unixTimeNow := currentTime.Unix()
				conn := connectionsCache[t][u].Connection
				if unixTimeNow > int64(user.ExpirationTime) {
					conn.Close()
					lock.Lock()
					delete(connectionsCache[t], u)
					lock.Unlock()
					fmt.Println("delete from cache", connectionsCache)
				}
			}
			if len(connectionsCache[t]) == 0 {
				lock.Lock()
				delete(connectionsCache, t)
				lock.Unlock()
			}
		}
	}
}
